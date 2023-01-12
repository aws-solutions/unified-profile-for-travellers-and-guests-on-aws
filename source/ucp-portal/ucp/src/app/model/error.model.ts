export class Error {
	constructor(
		public code: string,
		public transactionId: string,
		public errors: Array<any>,
		public error: any,
		public msg: string) { }

}