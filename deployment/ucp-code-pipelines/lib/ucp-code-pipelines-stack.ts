import { Stack, StackProps, SecretValue, CfnParameter, CfnCondition, Fn } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as kms from 'aws-cdk-lib/aws-kms';
import * as codebuild from 'aws-cdk-lib/aws-codebuild';
import * as codepipeline from 'aws-cdk-lib/aws-codepipeline';
import * as codepipeline_actions from 'aws-cdk-lib/aws-codepipeline-actions';
import * as sns from 'aws-cdk-lib/aws-sns';
import * as notifications from 'aws-cdk-lib/aws-codestarnotifications';
import * as subscriptions from 'aws-cdk-lib/aws-sns-subscriptions';
import * as iam from 'aws-cdk-lib/aws-iam';

import { NagSuppressions } from 'cdk-nag';
import * as tah_s3 from '../tah-cdk-common/s3';
import * as tah_core from '../tah-cdk-common/core';


const GO_VERSION = 1.18
const NODEJS_VERSION = 16
const UI_NODEJS_VERSION = 16
const RUBY_VERSION = 3.1
const PYTHON_VERSION = 3.9

export class UCPCodePipelinesStack extends Stack {

  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);
    const envName = this.node.tryGetContext("envName");
    const gitHubRepo = "unified-profile-for-travellers-and-guests-on-aws"

    //CloudFormatiion Input Parmetters to be provided by end user:
    const contactEmail = new CfnParameter(this, "contactEmail", {
      type: "String",
      allowedPattern: "^([a-zA-Z0-9_\\-\\.]+)@([a-zA-Z0-9_\\-\\.]+)\\.([a-zA-Z]{2,5})$",
      description: "Email address for the administrator"
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
    const buildFromUpstream = new CfnParameter(this, "buildFromUpstream", {
      type: "String",
      allowedValues: [
        "true",
        "false"
      ],
      default: "false",
      description: "If set to true, the solution will build from the upstream repository instead of your fork (AWS Internal)"
    });
    const branch = new CfnParameter(this, "branch", {
      type: "String",
      default: "main",
      description: "The git branch to build from"
    });

    const accessLogging = new tah_s3.AccessLogBucket(this, "ucp-pipeline-access-logging")
    const artifactBucket = new tah_s3.Bucket(this, "ucpArtifacts", accessLogging)
    const pipelineArtifactBucket = new tah_s3.Bucket(this, "ucpPipelineArtifacts", accessLogging)
    tah_core.Output.add(this, "accessLogging", accessLogging.bucketName)
    tah_core.Output.add(this, "artifactBucket", artifactBucket.bucketName)
    tah_core.Output.add(this, "pipelineArtifactBucket", pipelineArtifactBucket.bucketName)

    //TODO: add this role as parametter to avoid packing and Admin role with the solution
    const buildProjectRole = new iam.Role(this, 'buildRole', {
      assumedBy: new iam.ServicePrincipal('codebuild.amazonaws.com'),
      managedPolicies: [iam.ManagedPolicy.fromAwsManagedPolicyName("AdministratorAccess")]
    })

    let codeBuildKmsKey = new kms.Key(this, 'MyKey', {
      enableKeyRotation: true,
    });

    const infraBuild = new codebuild.PipelineProject(this, 'infraBuilProject' + envName, {
      projectName: "code-build-ucp-infra-" + envName,
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
              'npm install -g aws-cdk',
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
              'pwd && sh deploy.sh ' + envName + " " + artifactBucket.bucketName + " " + contactEmail.valueAsString
            ],
          },
        },
        artifacts: {
          "discard-path": "yes",
          files: [
            'source/ucp-infra/infra-config-' + envName + '.json',
          ],
        },
      }),
      environment: {
        buildImage: codebuild.LinuxBuildImage.AMAZON_LINUX_2_4,
      },
    });

    const lambdaBuild = new codebuild.PipelineProject(this, 'lambdaBuilProject' + envName, {
      projectName: "ucp-lambda-" + envName,
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
              'pwd && sh lbuild.sh ' + envName + " " + artifactBucket.bucketName
            ],
          },
        }
      }),
      environment: {
        buildImage: codebuild.LinuxBuildImage.AMAZON_LINUX_2_4,
      },
    });

    const syncLambdaBuild = new codebuild.PipelineProject(this, 'syncLambdaBuilProject' + envName, {
      projectName: "ucp-lambda-sync-" + envName,
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
              'cd source/ucp-sync',
              'pwd && sh lbuild.sh ' + envName + " " + artifactBucket.bucketName
            ],
          },
        }
      }),
      environment: {
        buildImage: codebuild.LinuxBuildImage.AMAZON_LINUX_2_4,
      },
    });

    const realTimeLambdaBuild = new codebuild.PipelineProject(this, 'realTimeLambdaBuild' + envName, {
      projectName: "ucp-lambda-real-time-" + envName,
      role: buildProjectRole,
      encryptionKey: codeBuildKmsKey,
      buildSpec: codebuild.BuildSpec.fromObject({
        version: '0.2',
        phases: {
          install: {
            "runtime-versions": {
              python: PYTHON_VERSION
            }
          },
          build: {
            commands: [
              'echo "Build and Deploy lambda Function"',
              'cd source/ucp-real-time-transformer',
              'pwd && sh deploy.sh ' + envName + " " + artifactBucket.bucketName
            ],
          },
        }
      }),
      environment: {
        buildImage: codebuild.LinuxBuildImage.AMAZON_LINUX_2_4,
      },
    });


    const firehoselambdaBuild = new codebuild.PipelineProject(this, 'streamLambdaBuilProject' + envName, {
      projectName: "ucp-stream-lambda-" + envName,
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
              'pwd && sh lbuild.sh ' + envName + " " + artifactBucket.bucketName
            ],
          },
        }
      }),
      environment: {
        buildImage: codebuild.LinuxBuildImage.AMAZON_LINUX_2_4,
      },
    });


    const onboardingTest = new codebuild.PipelineProject(this, 'testProject' + envName, {
      projectName: "ucp-test-" + envName,
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
              'pwd && sh ./test.sh ' + envName + " " + artifactBucket.bucketName
            ],
          },
        }
      }),
      environment: {
        buildImage: codebuild.LinuxBuildImage.AMAZON_LINUX_2_4,
      },
    });

    const feProject = new codebuild.PipelineProject(this, "adminPortal" + envName, {
      projectName: "ucp-admin-portal" + envName,
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
              'pwd && bash deploy.sh ' + envName + " " + artifactBucket.bucketName
            ],
          },
        }
      }),
      environment: {
        buildImage: codebuild.LinuxBuildImage.AMAZON_LINUX_2_4,
        computeType: codebuild.ComputeType.MEDIUM
      },
    });

    const etlProject = new codebuild.PipelineProject(this, 'etlDeployProject' + envName, {
      projectName: "ucp-etl-" + envName,
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
              'cd source/ucp-etls',
              'echo "Deploy ETL code"',
              'pwd && sh deploy.sh ' + envName + " " + artifactBucket.bucketName
            ],
          },
        }
      }),
      environment: {
        buildImage: codebuild.LinuxBuildImage.AMAZON_LINUX_2_4,
      },
    });

    const etlTestProject = new codebuild.PipelineProject(this, 'etlTestDeployProject' + envName, {
      projectName: "ucp-etl-test-" + envName,
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
              'cd source/ucp-etls',
              'echo "Testing ETLs"',
              'cd e2e',
              'sh test.sh ' + envName + " " + artifactBucket.bucketName,
            ],
          },
        }
      }),
      environment: {
        buildImage: codebuild.LinuxBuildImage.AMAZON_LINUX_2_4,
      },
    });

    const feTest = new codebuild.PipelineProject(this, 'feTestProject' + envName, {
      projectName: "ucp-fe-test-" + envName,
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
              "curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | sudo apt-key add -",
              "echo \"deb http://dl.google.com/linux/chrome/deb/ stable main\" >> /etc/apt/sources.list.d/google-chrome.list",
              'apt-get -y update',
              'apt-get -y install google-chrome-stable',
              'apt-get install -yq gconf-service xvfb libasound2 libatk1.0-0 libc6 libcairo2 libcups2 libdbus-1-3 libexpat1 libfontconfig1 libgcc1 libgconf-2-4 libgdk-pixbuf2.0-0 libglib2.0-0 libgtk-3-0 libnspr4 libpango-1.0-0 libpangocairo-1.0-0 libstdc++6 libx11-6 libx11-xcb1 libxcb1 libxcomposite1 libxcursor1 libxdamage1 libxext6 libxfixes3 libxi6 libxrandr2 libxrender1 libxss1 libxtst6 ca-certificates fonts-liberation libappindicator1 libnss3 lsb-release xdg-utils wget',
            ]
          },
          build: {
            commands: [
              'echo "Testing angular ui"',
              'pwd && sh ./test.sh ' + envName + " " + artifactBucket.bucketName
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
    const cdkBuildOutputLambdaSync = new codepipeline.Artifact('cdkBuildOutputLambdaSync');
    const cdkBuildOutputLambdaRealTime = new codepipeline.Artifact('cdkBuildOutputLambdaRealTime');
    const cdkBuildOutputEtl = new codepipeline.Artifact('CdkBuildOutputEtl');
    const cdkBuildOutputEtlTest = new codepipeline.Artifact('CdkBuildOutputEtlTest');
    const cdkBuildOutputFirehoseLambda = new codepipeline.Artifact('CdkBuildOutputFirehoseLambda');
    const cdkBuildOutputInfra = new codepipeline.Artifact('CdkBuildOutputInfra');
    const cdkBuildOutputTest = new codepipeline.Artifact('CdkBuildOutputTest');
    const feTestOutput = new codepipeline.Artifact('feTestOutput');
    const feOutput = new codepipeline.Artifact('feOutput');


    let stages: codepipeline.StageProps[] = []
    const buildFromUpstreamCondition = new CfnCondition(this, "buildFromUpstreamCondition", { expression: Fn.conditionEquals(buildFromUpstream.valueAsString, "true") });
    const branchEmptyCondition = new CfnCondition(this, "branchCondition", { expression: Fn.conditionEquals(branch.valueAsString, "") });
    let owner = Fn.conditionIf(buildFromUpstreamCondition.logicalId, "aws-solutions", gitHubUserName.valueAsString).toString();
    let branchName = Fn.conditionIf(branchEmptyCondition.logicalId, "main", branch.valueAsString).toString();
    //Source  stage
    stages.push({
      stageName: 'Source',
      actions: [
        new codepipeline_actions.GitHubSourceAction({
          actionName: 'GitHub_Source',
          repo: gitHubRepo,
          oauthToken: SecretValue.unsafePlainText(githubtoken.valueAsString),
          branch: branchName,
          owner: owner,
          //we pevent automated pull as we use code pipeine as a deployment mecanism
          trigger: codepipeline_actions.GitHubTrigger.NONE,
          output: sourceOutput,
        }),
      ],
    })
    //Build  stage
    stages.push({
      stageName: 'Build',
      actions: [
        new codepipeline_actions.CodeBuildAction({
          actionName: 'deployEtlCode',
          project: etlProject,
          input: sourceOutput,
          runOrder: 1,
          outputs: [cdkBuildOutputEtl],
        }),
        new codepipeline_actions.CodeBuildAction({
          actionName: 'buildLambdaCode',
          project: lambdaBuild,
          input: sourceOutput,
          runOrder: 2,
          outputs: [cdkBuildOutputLambda],
        }),
        new codepipeline_actions.CodeBuildAction({
          actionName: 'buildSyncLambdaCode',
          project: syncLambdaBuild,
          input: sourceOutput,
          runOrder: 2,
          outputs: [cdkBuildOutputLambdaSync],
        }),
        new codepipeline_actions.CodeBuildAction({
          actionName: 'buildRealTimeLambdaCode',
          project: realTimeLambdaBuild,
          input: sourceOutput,
          runOrder: 2,
          outputs: [cdkBuildOutputLambdaRealTime],
        }),
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
      stageName: 'EndToEndTests',
      actions: [
        new codepipeline_actions.CodeBuildAction({
          actionName: 'apiTests',
          project: onboardingTest,
          input: sourceOutput,
          outputs: [cdkBuildOutputTest],
          runOrder: 1,
        }),
        new codepipeline_actions.CodeBuildAction({
          actionName: 'runFeTests',
          project: feTest,
          input: sourceOutput,
          runOrder: 1,
          outputs: [feTestOutput],
        }),
        new codepipeline_actions.CodeBuildAction({
          actionName: 'testEtlCode',
          project: etlTestProject,
          input: sourceOutput,
          runOrder: 1,
          outputs: [cdkBuildOutputEtlTest],
        })
      ],
    })
    //Deploy Stages
    let deployStage: codepipeline.StageProps = {
      stageName: 'Deploy',
      actions: [],
    }
    if (deployStage.actions) {
      deployStage.actions.push(new codepipeline_actions.CodeBuildAction({
        actionName: 'deployFrontEnd',
        project: feProject,
        input: sourceOutput,
        runOrder: 2,
        outputs: [feOutput],
      }))
    }
    stages.push(deployStage)



    let pipeline = new codepipeline.Pipeline(this, 'ucpPipeline' + envName, {
      pipelineName: "ucp-" + envName,
      stages: stages,
      crossAccountKeys: false,
      artifactBucket: pipelineArtifactBucket,
    });

    ///////////////////////////
    // Notification on Complete
    ////////////////////////////
    let snsEncryptionKey = new kms.Key(this, 'SnsEncryptionKey', {
      enableKeyRotation: true,
    });
    const topic = new sns.Topic(this, 'Topic', {
      displayName: 'AWS T&H Solution CICD Pipeline Events ' + envName,
      masterKey: snsEncryptionKey,
    });
    topic.addSubscription(new subscriptions.EmailSubscription(contactEmail.valueAsString));

    const rule = new notifications.NotificationRule(this, 'NotificationRule', {
      source: pipeline,
      events: [
        'codepipeline-pipeline-pipeline-execution-failed',
        'codepipeline-pipeline-pipeline-execution-canceled',
        'codepipeline-pipeline-pipeline-execution-started',
        'codepipeline-pipeline-pipeline-execution-resumed',
        'codepipeline-pipeline-pipeline-execution-succeeded',
        'codepipeline-pipeline-pipeline-execution-superseded'
      ],
      targets: [topic],
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
    NagSuppressions.addResourceSuppressions(buildProjectRole, [
      {
        id: 'AwsSolutions-IAM4',
        reason: 'Administrator role permission for build role required'
      },
    ], true);
  }

}






