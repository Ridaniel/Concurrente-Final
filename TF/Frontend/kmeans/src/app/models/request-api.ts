import { Car } from './car';

export class RequestApi {
    car: Car;
    k: number;
    constructor(car: Car,k: number){
        this.car= car;
        this.k = k;
    }
}
