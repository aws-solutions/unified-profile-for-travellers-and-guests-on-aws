export class MetadataItem {
	public id: string;
	public name: string;
	public description: string;
	public key: string;
	public allowMultipleInstances: boolean;
	public type: string;
	public readAccess: string;
	public writeAccess: string;
	public category: string;
	public subCategory: string;
	public parent: string;
	public needsTranslation: boolean;
	public needsApproval: boolean;
	constructor() { };
}