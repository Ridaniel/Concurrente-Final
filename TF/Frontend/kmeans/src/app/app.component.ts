import { Component } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { HelperComponent } from './helper/helper.component';
import { Car } from './models/car';
import { CarServiceService } from './services/car-service.service';
import { ResultComponent } from './result/result.component';
import { RequestApi } from './models/request-api';
import { User } from './models/user';
import { Region } from './models/region';
import { range } from 'rxjs';
import { Data } from './models/data';
import { connectableObservableDescriptor } from 'rxjs/internal/observable/ConnectableObservable';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})

export class AppComponent  {

  user: User = new User();
  regiones: Region[];
  region: Region;

  value: number[];

  constructor(private dialog: MatDialog, private carService: CarServiceService){
    this.colocarRegiones();
  }

  openHelper(): void{
    const dialogRef = this.dialog.open(HelperComponent);
  }
  
  diagnosticar(){
    this.user.insuficienciaRespiratoria ? this.user.insuficienciaRespiratoria = 1 : this.user.insuficienciaRespiratoria = 0;
    this.user.neumonia ? this.user.neumonia = 1 : this.user.neumonia = 0;
    this.user.viajo ? this.user.viajo = 1 : this.user.viajo = 0;
    this.user.sexo = Number(this.user.sexo);


    this.value = new Array<number>();
    this.value.push(this.user.edad);
    this.value.push(this.user.sexo);
    this.value.push(this.user.region);
    this.value.push(this.user.viajo);
    this.value.push(this.user.insuficienciaRespiratoria);
    this.value.push(this.user.neumonia);
    this.value.push(0);
    this.value.push(0);

    var requestData =new Data(this.value);

    
    console.log(requestData);

    this.carService.metodo(requestData).subscribe(
      result => {
        console.log("Enviar");
        console.log(result);
        //const dialogRef = this.dialog.open(ResultComponent, {data:{result: result}});
      },
      err => {alert("Ocurrio un error, no se pudo obtener los datos");console.error(err);}
    );
    
  }

  colocarRegiones(){

    this.regiones = new Array<Region>();
    let listaRegiones = ["Lima", 0, "Piura", 1, "La Libertad", 2, "Cusco", 3, "Lambayeque", 4, "Callao", 5,"Ancash", 6, "Loreto", 7, "Tumbes", 8, "San Martin", 9, "Arequipa", 10, "Junin", 11 ];

    for (var i = 0; i < listaRegiones.length; i++) {
      this.region = new Region();
      this.region.nombre = String(listaRegiones[i]);
      i++;
      this.region.valor = Number(listaRegiones[i]);
      this.regiones.push(this.region);
    }
  }
}
