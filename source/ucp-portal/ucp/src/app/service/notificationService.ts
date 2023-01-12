import { Injectable } from '@angular/core';
import { MatSnackBar, MatSnackBarVerticalPosition } from '@angular/material/snack-bar';


import * as _ from 'lodash';


@Injectable()
export class NotificationService {
  durationInSeconds = 10;
  verticalPosition: MatSnackBarVerticalPosition = 'top';
  actionName = "dismiss"


  constructor(private _snackBar: MatSnackBar) {

  }

  showConfirmation(msg: string) {
    console.log("Show confirmation: " + msg)
    this._snackBar.open(msg, this.actionName, {
      duration: this.durationInSeconds * 1000,
      verticalPosition: this.verticalPosition,
      panelClass: "notif-confirmation"
    });
  }

  showInfo(msg: string) {
    console.log("Show info: " + msg)
    this._snackBar.open(msg, this.actionName, {
      duration: this.durationInSeconds * 1000,
      verticalPosition: this.verticalPosition,
      panelClass: "notif-info"
    });
  }

  showError(msg: string) {
    console.log("Show error: " + msg)
    this._snackBar.open(msg, this.actionName, {
      duration: this.durationInSeconds * 1000,
      verticalPosition: this.verticalPosition,
      panelClass: "notif-error"
    });
  }
  showWarning(msg: string) {
    console.log("Show warning: " + msg)
    this._snackBar.open(msg, this.actionName, {
      duration: this.durationInSeconds * 1000,
      verticalPosition: this.verticalPosition,
      panelClass: "notif-warning"
    });
  }

}

