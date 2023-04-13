export class PaginationOptions {
    public page: number
    public pageSize: number
    public objectType?: string
    constructor(page: number, pageSize: number, objectType: string) {
        this.page = page
        this.pageSize = pageSize
        this.objectType = objectType
    }

}