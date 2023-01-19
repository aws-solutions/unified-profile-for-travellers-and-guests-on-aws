import { Stack, StackProps, SecretValue, CfnParameter } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as kms from 'aws-cdk-lib/aws-kms';
import * as codebuild from 'aws-cdk-lib/aws-codebuild';
import * as codepipeline from 'aws-cdk-lib/aws-codepipeline';
import * as codepipeline_actions from 'aws-cdk-lib/aws-codepipeline-actions';
import { NagSuppressions } from 'cdk-nag';
import * as tah_s3 from '../tah-cdk-common/s3';
import * as tah_core from '../tah-cdk-common/core';
import * as tah_codepipeline from '../tah-cdk-common/codepipeline';


const GO_VERSION = 1.18
const NODEJS_VERSION = 16
const UI_NODEJS_VERSION = 16
const RUBY_VERSION = 3.1

export class UCPCodePipelinesStack extends Stack {

  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);

    const gitHubRepo = "unified-profile-for-travellers-and-guests-on-aws"

    //CloudFormatiion Input Parmetters to be provided by end user:
    const contactEmail = new CfnParameter(this, "contactEmail", {
      type: "String",
      allowedPattern: "^([a-zA-Z0-9_\\-\\.]+)@([a-zA-Z0-9_\\-\\.]+)\\.([a-zA-Z]{2,5})$",
      description: "Email address for the administrator"
    });
    const envNameVal = new CfnParameter(this, "environment", {
      type: "String",
      allowedValues: ["dev", "int", "staging", "prod"],
      default: "int",
      description: "Your environment name. Change to a unique name only if deploy the stack multiple times in the same region and account."
    });
    const gitHubUserName = new CfnParameter(this, "gitHubUserName", {
      type: "String",
      allowedPattern: ".+",
      description: "Your github user name (see pre-deployment steps)"
    });

    const githubtoken = new CfnParameter(this, "githubtoken", {
      type: "String",
      allowedPattern: ".+",
      noEcho: true,
      description: "Your github Personal access tokens allowing access to the forked repository (see pre-deployment steps)"
    });

    const accessLogging = new tah_s3.AccessLogBucket(this, "aws-tah-industry-connector-pipeline-access-logging")
    const artifactBucket = new tah_s3.Bucket(this, "ucpArtifacts", accessLogging)
    const pipelineArtifactBucket = new tah_s3.Bucket(this, "ucpPipelineArtifacts", accessLogging)
    tah_core.Output.add(this, "accessLogging", accessLogging.bucketName)
    tah_core.Output.add(this, "artifactBucket", artifactBucket.bucketName)
    tah_core.Output.add(this, "pipelineArtifactBucket", pipelineArtifactBucket.bucketName)
    //build role with admin access
    const buildProjectRole = new tah_codepipeline.BuildRole(this, 'buildRole')

    let codeBuildKmsKey = new kms.Key(this, 'MyKey', {
      enableKeyRotation: true,
    });

    const infraBuild = new codebuild.PipelineProject(this, 'infraBuilProject', {
      projectName: "code-build-ucp-infra-" + envNameVal.valueAsString,
      role: buildProjectRole,
      encryptionKey: codeBuildKmsKey,
      buildSpec: codebuild.BuildSpec.fromObject({
        version: '0.2',
        phases: {
          install: {
            "runtime-versions": {
              nodejs: NODEJS_VERSION,
              ruby: RUBY_VERSION
            },
            commands: [
              'echo "CodeBuild is running in $AWS_REGION" && aws configure set region $AWS_REGION',
              'npm install -g aws-cdk@2.60.0',
              'npm -g install typescript@4.2.2',
              'cdk --version',
              'gem install cfn-nag',
              'cd source/ucp-infra',
              'npm install'
            ]
          },
          build: {
            commands: [
              'echo "Build and Deploy Infrastructure"',
              'pwd && sh deploy.sh ' + envNameVal.valueAsString + " " + artifactBucket.bucketName + " " + contactEmail.valueAsString
            ],
          },
        },
        artifacts: {
          "discard-path": "yes",
          files: [
            'source/ucp-infra/infra-config-' + envNameVal.valueAsString + '.json',
          ],
        },
      }),
      environment: {
        buildImage: codebuild.LinuxBuildImage.AMAZON_LINUX_2_4,
      },
    });

    const lambdaBuild = new codebuild.PipelineProject(this, 'lambdaBuilProject', {
      projectName: "ucp-lambda-" + envNameVal.valueAsString,
      role: buildProjectRole,
      encryptionKey: codeBuildKmsKey,
      buildSpec: codebuild.BuildSpec.fromObject({
        version: '0.2',
        phases: {
          install: {
            "runtime-versions": {
              golang: GO_VERSION
            }
          },
          build: {
            commands: [
              'echo "Build and Deploy lambda Function"',
              'cd source/ucp-backend',
              'pwd && sh lbuild.sh ' + envNameVal.valueAsString + " " + artifactBucket.bucketName
            ],
          },
        }
      }),
      environment: {
        buildImage: codebuild.LinuxBuildImage.AMAZON_LINUX_2_4,
      },
    });
    const firehoselambdaBuild = new codebuild.PipelineProject(this, 'streamLambdaBuilProject', {
      projectName: "ucp-stream-lambda-" + envNameVal.valueAsString,
      role: buildProjectRole,
      encryptionKey: codeBuildKmsKey,
      buildSpec: codebuild.BuildSpec.fromObject({
        version: '0.2',
        phases: {
          install: {
            "runtime-versions": {
              golang: GO_VERSION
            }
          },
          build: {
            commands: [
              'echo "Build and Deploy lambda Function"',
              'cd source/ucp-stream-processor',
              'pwd && sh lbuild.sh ' + envNameVal.valueAsString + " " + artifactBucket.bucketName
            ],
          },
        }
      }),
      environment: {
        buildImage: codebuild.LinuxBuildImage.AMAZON_LINUX_2_4,
      },
    });





    const onboardingTest = new codebuild.PipelineProject(this, 'testProject', {
      projectName: "ucp-test-" + envNameVal.valueAsString,
      role: buildProjectRole,
      encryptionKey: codeBuildKmsKey,
      buildSpec: codebuild.BuildSpec.fromObject({
        version: '0.2',
        phases: {
          install: {
            "runtime-versions": {
              nodejs: NODEJS_VERSION
            },
            commands: [
              "npm install -g newman@5.2.2"
            ]
          },
          build: {
            commands: [
              'echo "Testing Api using newman"',
              'cd source/ucp-backend/e2e',
              'pwd && sh ./test.sh ' + envNameVal.valueAsString + " " + artifactBucket.bucketName
            ],
          },
        }
      }),
      environment: {
        buildImage: codebuild.LinuxBuildImage.AMAZON_LINUX_2_4,
      },
    });

    const feProject = new codebuild.PipelineProject(this, "adminPortal", {
      projectName: "ucp-admin-portal",
      role: buildProjectRole,
      encryptionKey: codeBuildKmsKey,
      buildSpec: codebuild.BuildSpec.fromObject({
        version: '0.2',
        phases: {
          install: {
            "runtime-versions": {
              nodejs: UI_NODEJS_VERSION
            },
            commands: [
              'echo "CodeBuild is running in $AWS_REGION" && aws configure set region $AWS_REGION',
              'cd source/ucp-portal/ucp',
              'npm install -g @angular/cli',
              'npm install --legacy-peer-deps'
            ]
          },
          build: {
            commands: [
              'echo "Build and Deploy Front end"',
              'pwd && bash deploy.sh ' + envNameVal.valueAsString + " " + artifactBucket.bucketName
            ],
          },
        }
      }),
      environment: {
        buildImage: codebuild.LinuxBuildImage.AMAZON_LINUX_2_4,
        computeType: codebuild.ComputeType.MEDIUM
      },
    })

    const feTest = new codebuild.PipelineProject(this, 'feTestProject', {
      projectName: "ucp-fe-test-" + envNameVal.valueAsString,
      role: buildProjectRole,
      encryptionKey: codeBuildKmsKey,
      buildSpec: codebuild.BuildSpec.fromObject({
        version: '0.2',
        phases: {
          install: {
            "runtime-versions": {
              nodejs: NODEJS_VERSION
            },
            commands: [
              'cd source/ucp-portal/ucp',
              'cd e2e',
              'echo "5-Install Angular"',
              'npm install -g @angular/cli',
              'npm install --legacy-peer-deps',
              'echo "2-Install headless Chrome"',
              'curl -sS -o - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add -',
              'deb http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google-chrome.list',
              'apt-get -y update',
              'apt-get -y install google-chrome-stable',
              'apt-get install -yq gconf-service xvfb libasound2 libatk1.0-0 libc6 libcairo2 libcups2 libdbus-1-3 libexpat1 libfontconfig1 libgcc1 libgconf-2-4 libgdk-pixbuf2.0-0 libglib2.0-0 libgtk-3-0 libnspr4 libpango-1.0-0 libpangocairo-1.0-0 libstdc++6 libx11-6 libx11-xcb1 libxcb1 libxcomposite1 libxcursor1 libxdamage1 libxext6 libxfixes3 libxi6 libxrandr2 libxrender1 libxss1 libxtst6 ca-certificates fonts-liberation libappindicator1 libnss3 lsb-release xdg-utils wget',
            ]
          },
          build: {
            commands: [
              'echo "Testing angular ui"',
              'pwd && sh ./test.sh ' + envNameVal.valueAsString + " " + artifactBucket.bucketName
            ],
          },
        }
      }),
      environment: {
        buildImage: codebuild.LinuxBuildImage.STANDARD_6_0,
      },
    });

    //Output Artifacts
    const sourceOutput = new codepipeline.Artifact();
    const cdkBuildOutputLambda = new codepipeline.Artifact('CdkBuildOutputLambda');
    const cdkBuildOutputFirehoseLambda = new codepipeline.Artifact('CdkBuildOutputFirehoseLambda');
    const cdkBuildOutputInfra = new codepipeline.Artifact('CdkBuildOutputInfra');
    const cdkBuildOutputTest = new codepipeline.Artifact('CdkBuildOutputTest');
    const feTestOutput = new codepipeline.Artifact('feTestOutput');
    const feOutput = new codepipeline.Artifact('feOutput');


    let stages: codepipeline.StageProps[] = []
    let branchName = "main"
    //we isoluate productino to a dedicated branch to ensure a proper productino deployment process
    //requireing a pull request formo main to production branch
    if (envNameVal.valueAsString === 'prod') {
      branchName = "production"
    }
    //Source  stage
    stages.push({
      stageName: 'Source',
      actions: [
        new codepipeline_actions.GitHubSourceAction({
          actionName: 'GitHub_Source',
          repo: gitHubRepo,
          oauthToken: SecretValue.unsafePlainText(githubtoken.valueAsString),
          branch: branchName,
          owner: gitHubUserName.valueAsString,
          output: sourceOutput,
        }),
      ],
    })
    //Build  stage
    stages.push({
      stageName: 'Build',
      actions: [
        new codepipeline_actions.CodeBuildAction({
          actionName: 'buildLambdaCode',
          project: lambdaBuild,
          input: sourceOutput,
          runOrder: 2,
          outputs: [cdkBuildOutputLambda],
        }),
        /*new codepipeline_actions.CodeBuildAction({
          actionName: 'buildFirehoseLambdaCode',
          project: firehoselambdaBuild,
          input: sourceOutput,
          runOrder: 2,
          outputs: [cdkBuildOutputFirehoseLambda],
        }),*/
        new codepipeline_actions.CodeBuildAction({
          actionName: 'deployInfra',
          project: infraBuild,
          input: sourceOutput,
          runOrder: 3,
          outputs: [cdkBuildOutputInfra],
        }),
      ],
    })
    //Test Stage
    stages.push({
      stageName: 'Test',
      actions: [
        new codepipeline_actions.CodeBuildAction({
          actionName: 'e2eTesting',
          project: onboardingTest,
          input: sourceOutput,
          outputs: [cdkBuildOutputTest],
        }),
      ],
    })
    //Deploy Stages
    let deployStage: codepipeline.StageProps = {
      stageName: 'Deploy',
      actions: [],
    }
    if (deployStage.actions) {
      deployStage.actions.push(new codepipeline_actions.CodeBuildAction({
        actionName: 'runFeTests',
        project: feTest,
        input: sourceOutput,
        runOrder: 1,
        outputs: [feTestOutput],
      }))
      deployStage.actions.push(new codepipeline_actions.CodeBuildAction({
        actionName: 'deployFrontEnd',
        project: feProject,
        input: sourceOutput,
        runOrder: 2,
        outputs: [feOutput],
      }))

    }
    stages.push(deployStage)



    let pipeline = new codepipeline.Pipeline(this, 'ucpPipeline', {
      pipelineName: "ucp-" + envNameVal.valueAsString,
      stages: stages,
      crossAccountKeys: false,
      artifactBucket: pipelineArtifactBucket,
    });
    NagSuppressions.addResourceSuppressions(pipeline.role, [
      {
        id: 'AwsSolutions-IAM5',
        reason: 'Role generated by CodePipeline CDK'
      },
    ], true);
    NagSuppressions.addResourceSuppressions(buildProjectRole, [
      {
        id: 'AwsSolutions-IAM5',
        reason: 'Administrator role permission for build role required'
      },
    ], true);

  }

}






