// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import { CdkBase, CdkBaseProps } from '../cdk-base';

export type UptVpcProps =
    | {
          isByoVpcTemplate: false;
      }
    | {
          isByoVpcTemplate: true;
          vpcIdParam: cdk.CfnParameter;
          vpcCidrBlockParam: cdk.CfnParameter;
          privateSubnetId1Param: cdk.CfnParameter;
          privateSubnetId2Param: cdk.CfnParameter;
          privateSubnetId3Param: cdk.CfnParameter;
          privateSubnetRouteTableId1Param: cdk.CfnParameter;
          privateSubnetRouteTableId2Param: cdk.CfnParameter;
          privateSubnetRouteTableId3Param: cdk.CfnParameter;
      };

export class UptVpc extends CdkBase {
    public uptVpc: ec2.IVpc;
    private props: UptVpcProps;

    constructor(cdkProps: CdkBaseProps, props: UptVpcProps) {
        super(cdkProps);
        this.props = props;

        // Import or create VPC
        if (this.props.isByoVpcTemplate) {
            this.uptVpc = this.importVpc();
        } else {
            this.uptVpc = this.createVpc();
        }

        // Add required resources
        this.addSecretsManagerInterfaceEndpoint();
        this.addDynamoGatewayEndpoint();
    }

    /**
     * Returns private subnet IDs for the VPC
     */
    public getPrivateSubnetIds(): string[] {
        if (this.props.isByoVpcTemplate) {
            return [
                this.props.privateSubnetId1Param.valueAsString,
                this.props.privateSubnetId2Param.valueAsString,
                this.props.privateSubnetId3Param.valueAsString
            ];
        }

        return this.uptVpc.selectSubnets().subnetIds;
    }

    /**
     * Imports a customer's existing VPC to be used by UPT
     * @returns VPC
     */
    private importVpc(): ec2.IVpc {
        if (!this.props.isByoVpcTemplate) {
            throw new Error('Import is only valid when using BYO VPC template');
        }

        const vpc = ec2.Vpc.fromVpcAttributes(this.stack, 'ImportedVPC', {
            vpcId: this.props.vpcIdParam.valueAsString,
            availabilityZones: cdk.Fn.getAzs(this.stack.region),
            privateSubnetIds: [
                this.props.privateSubnetId1Param.valueAsString,
                this.props.privateSubnetId2Param.valueAsString,
                this.props.privateSubnetId3Param.valueAsString
            ],
            privateSubnetRouteTableIds: [
                this.props.privateSubnetRouteTableId1Param.valueAsString,
                this.props.privateSubnetRouteTableId2Param.valueAsString,
                this.props.privateSubnetRouteTableId3Param.valueAsString
            ],
            vpcCidrBlock: this.props.vpcCidrBlockParam.valueAsString
        });

        return vpc;
    }

    /**
     * Creates a VPC to be used by UPT
     * @returns VPC
     */
    private createVpc(): ec2.IVpc {
        const cidrBlock = '10.0.0.0/16'; // NOSONAR (typescript:S1313)
        const vpc = new ec2.Vpc(this.stack, 'ucpDbVpc', {
            vpcName: 'ucp-vpc-' + this.envName,
            ipAddresses: ec2.IpAddresses.cidr(cidrBlock),
            maxAzs: 3,
            subnetConfiguration: [
                { name: 'public', subnetType: ec2.SubnetType.PUBLIC, cidrMask: 24 },
                { name: 'egress', subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS, cidrMask: 24 }
            ]
        });

        return vpc;
    }

    /**
     * Adds a Secrets Manager interface endpoint to the VPC
     * see more: https://docs.aws.amazon.com/whitepapers/latest/aws-privatelink/what-are-vpc-endpoints.html
     */
    private addSecretsManagerInterfaceEndpoint(): void {
        this.uptVpc.addInterfaceEndpoint('SecretsManagerEndpoint', {
            service: ec2.InterfaceVpcEndpointAwsService.SECRETS_MANAGER
        });

        this.uptVpc.node.children.forEach(child => {
            if (child.node.defaultChild instanceof ec2.CfnVPCEndpoint) {
                child.node.children.forEach(grandchild => {
                    if (grandchild.node.defaultChild instanceof ec2.CfnSecurityGroup) {
                        (grandchild.node.defaultChild as ec2.CfnSecurityGroup).cfnOptions.metadata = {
                            cfn_nag: {
                                rules_to_suppress: [
                                    {
                                        id: 'W5',
                                        reason: 'Not applicable'
                                    },
                                    {
                                        id: 'W40',
                                        reason: 'Not applicable'
                                    }
                                ]
                            }
                        };
                    }
                });
            }
        });
    }

    /**
     * Adds a DynamoDB Gateway endpoint to the VPC
     * see more: https://docs.aws.amazon.com/whitepapers/latest/aws-privatelink/what-are-vpc-endpoints.html
     */
    private addDynamoGatewayEndpoint(): void {
        this.uptVpc.addGatewayEndpoint('DynamoEndpoint', {
            service: ec2.GatewayVpcEndpointAwsService.DYNAMODB
        });
    }
}
