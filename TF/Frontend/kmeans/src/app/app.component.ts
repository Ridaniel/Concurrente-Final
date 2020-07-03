import { Component } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { CarServiceService } from './services/car-service.service';
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
  text: string;
  textResp: string;

  resp: boolean;

  constructor(private dialog: MatDialog, private carService: CarServiceService){
    this.colocarRegiones();
    this.text = "";
    this.resp = false;
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
    
    this.carService.metodo(requestData).subscribe(
      result => {
        this.text = "El porcentaje de peligro para la persona es: " + Number((result.riesgo * 100)).toFixed(2) + "%";
        this.resp = result.infectado;
        if(result.infectado == null){
          this.textResp = "Negativo a Covid-19";
        }else {
          this.textResp = "Positivo a Covid-19";
        }
      },
      err => {alert("Ocurrio un error, no se pudo obtener los datos");console.error(err);}
    );

    this.user = new User();
    
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
