import { MetadataItem } from "./metadataItem.model"

export class Category {
    public name: string
    public subcategories: Array<SubCategory>
    constructor() { };
}

export class SubCategory {
    public name: string
    public items: Array<MetadataItem>
    constructor() { };
}