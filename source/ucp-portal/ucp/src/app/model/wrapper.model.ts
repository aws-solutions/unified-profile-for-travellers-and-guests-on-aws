import { Error } from './error.model';

export class Wrapper {
  constructor(
    public response: Object,
    public status: number,
    public error: Error) { }
}