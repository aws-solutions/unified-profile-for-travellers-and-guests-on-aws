import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { AuthenticationDetails, CognitoUser, CognitoUserPool } from 'amazon-cognito-identity-js';
import { Configuration } from '../app.constants';
import { FormControl } from '@angular/forms';
import { SessionService } from '../service/sessionService'
import { AuthContext } from '../model/auth.model'
import { NotificationService } from '../service/notificationService';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css']
})
export class LoginComponent implements OnInit {
  username = new FormControl('');
  password = new FormControl('');
  password2 = new FormControl('');
  sessionUserAttributes: any;
  cognitoUser: any;
  showConfirmPassword = false;
  constructor(private router: Router, private config: Configuration, private sessionService: SessionService, private notif: NotificationService) { }

  ngOnInit() {
  }

  signIn() {
    let authenticationDetails = new AuthenticationDetails({
      Username: this.username.value,
      Password: this.password.value,
    });
    let poolData = {
      UserPoolId: this.config.AWS_USER_POOL_ID, // Your user pool id here
      ClientId: this.config.AWS_CLIENT_ID // Your client id here
    };

    let userPool = new CognitoUserPool(poolData);
    let userData = { Username: this.username.value, Pool: userPool };
    this.cognitoUser = new CognitoUser(userData);
    console.log("Auth ", authenticationDetails)
    this.cognitoUser.authenticateUser(authenticationDetails, {
      onSuccess: (session) => {
        console.log(`Signed in user ${this.cognitoUser.getUsername()}. Sessiong valid?: `, session.isValid())
        console.log(`Session: `, session)

        let authCtx = new AuthContext()
        authCtx.token = session.getAccessToken().getJwtToken()
        authCtx.login = this.cognitoUser.getUsername()
        console.log(`[AuthService] Setting auth context in storage: `, authCtx)
        this.sessionService.create(authCtx);
        setTimeout(() => { this.router.navigate(["home"]) }, 500);


      },
      newPasswordRequired: (userAttributes, requiredAttributes) => {
        this.notif.showInfo("Please Change your Password")
        this.password.reset();
        this.password2.reset();
        this.showConfirmPassword = true
        console.log(userAttributes)
        console.log(requiredAttributes)
        delete userAttributes.email_verified;
        delete userAttributes.email;
        this.sessionUserAttributes = userAttributes;
      },
      onFailure: (err) => {
        this.notif.showError(err.message || JSON.stringify(err));
      },
    });


  }

  updatePassword() {
    console.log("New ", this.password.value)
    console.log("Attributes: ", this.sessionUserAttributes)
    this.cognitoUser.completeNewPasswordChallenge(this.password.value, this.sessionUserAttributes, {
      onSuccess: (result) => {
        console.log(result)
        this.showConfirmPassword = false
        this.password.reset();
        this.notif.showConfirmation("Successfully changed password. Please log in")

      },
      onFailure: (err) => {
        console.log(err)
        this.notif.showError(err.message || JSON.stringify(err));

      }

    })
  }

}
