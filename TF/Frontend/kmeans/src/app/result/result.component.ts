import { Component, OnInit, Inject } from '@angular/core';
import { ResultApi } from '../models/result-api';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { element } from 'protractor';

@Component({
  selector: 'app-result',
  templateUrl: './result.component.html',
  styleUrls: ['./result.component.css']
})
export class ResultComponent implements OnInit {
  
  class: string = "";
  constructor(public dialogRef: MatDialogRef<ResultComponent>,@Inject(MAT_DIALOG_DATA) public data: {result: ResultApi[]}) { }

  ngOnInit(): void {
    this.calcClass();

  }

  calcClass(){
    var arr: number[] = new Array(this.data.result.length);
    this.data.result.forEach((x,index) => {
      arr[index] = 0;
      this.data.result.forEach(e=> {
            if(e.class == x.class){
              arr[0] = arr[0]++;
            }
        });
    });
    
    var a = Math.max(...arr);
    this.class = this.data.result[a].class;
    console.log(this.data.result,this.class);
    }
}
