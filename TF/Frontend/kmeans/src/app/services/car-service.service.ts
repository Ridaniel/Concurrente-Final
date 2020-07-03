import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Data } from '../models/data';
import { Pronostico } from '../models/pronostico';

const headers = new HttpHeaders({'Content-Type':  'application/json'})
const httpOptions = { headers: headers};
@Injectable({
  providedIn: 'root'
})
export class CarServiceService {
  readonly apiUrl: string = "http://localhost:3000/foo";
  constructor(private http: HttpClient) { 
  }
  metodo(data: Data): Observable<Pronostico> {
    return this.http.post<Pronostico>(this.apiUrl,data,httpOptions);
  }

}
