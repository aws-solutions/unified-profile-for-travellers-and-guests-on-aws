import { Component } from '@angular/core';
import { FormGroup, FormArray, FormControl, Validators, AbstractControl } from '@angular/forms';
import { LoaderService } from './service/loaderService';
import { faSpinner } from '@fortawesome/free-solid-svg-icons';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent {
  title = 'rep';
  faSpinner = faSpinner
  public itemForm = new FormGroup({})

  constructor(public _loader: LoaderService) {

  }
  public isLoading() {
    return this._loader.isLoading();
  }
}
