export class LoginRq {
	public login: string;
	public deviceToken: string;
	public phone: string;
	public password: string;
	public application: string;
	constructor() { };
}

export class SignUpRq {
	public login: string;
	public email: string;
	public phone: string;
	public password: string;
	public password2: string;
	public application: string;
	constructor() { };
}

export class ActivationRq {
	public login: string;
	public code: string;
	public application: string;
	constructor() { };
}

export class AuthContext {
	public token: string;
	public idToken: string;
	public login: string;
	public application: string;
	constructor() { };
}