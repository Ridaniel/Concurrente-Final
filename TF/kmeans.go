package main

//Copyright (c) 2014, Bugra Akyildiz

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"cors-master"
	"bytes"
	"strconv"
	"mux/mux-master"
	"strings"
	"time"
)
var means []Observation

// 2-norm distance (l_2 distance)
func EuclideanDistance(firstVector, secondVector []float64) float64 {
	distance := 0.
	for ii := range firstVector {
		distance += (firstVector[ii] - secondVector[ii]) * (firstVector[ii] - secondVector[ii])
	}
	return math.Sqrt(distance)
}

// Observation: Data Abstraction for an N-dimensional
// observation
type Observation []float64

// Abstracts the Observation with a cluster number
// Update and computeation becomes more efficient
type ClusteredObservation struct {
	ClusterNumber int
	Observation
}

// Distance Function: To compute the distanfe between observations
type DistanceFunction func(first, second []float64) (float64, error)


type Data struct {
	Values Observation `json:"values,omitempty"`
}

type Pronostic struct{
   Infectado bool  `json:"infectado,omitempty"`
   Riesgo float64  `json:"riesgo,omitempty"`
} 

type Diagnostico struct {
   Edad float64 `json:"edad,omitempty"`
   Sexo float64 `json:"sexo,omitempty"`
   Region float64 `json:"region,omitempty"`
   Viaje float64 `json:"viaje,omitempty"`
   InsuficienciaRespiratoria float64 `json:"insuficienciaRespiratoria,omitempty"`
   Neumonia float64 `json:"neumonia,omitempty"`
   Infectado bool `json:"infectado,omitempty"`
   Riesgo float64 `json:"riesgo,omitempty"`
}

// Summation of two vectors
func (observation Observation) Add(otherObservation Observation) {
	for ii, jj := range otherObservation {
		observation[ii] += jj
	}
}

// Multiplication of a vector with a scalar
func (observation Observation) Mul(scalar float64) {
	for ii := range observation {
		observation[ii] *= scalar
	}
}

// Dot Product of Two vectors
func (observation Observation) InnerProduct(otherObservation Observation) {
	for ii := range observation {
		observation[ii] *= otherObservation[ii]
	}
}

// Outer Product of two arrays
// TODO: Need to be tested
func (observation Observation) OuterProduct(otherObservation Observation) [][]float64 {
	result := make([][]float64, len(observation))
	for ii := range result {
		result[ii] = make([]float64, len(otherObservation))
	}
	for ii := range result {
		for jj := range result[ii] {
			result[ii][jj] = observation[ii] * otherObservation[jj]
		}
	}
	return result
}

// Find the closest observation and return the distance
// Index of observation, distance
func near(p ClusteredObservation, mean []Observation) (int, float64) {
	indexOfCluster := 0
	minSquaredDistance := EuclideanDistance(p.Observation, mean[0])
	for i := 1; i < len(mean); i++ {
		squaredDistance := EuclideanDistance(p.Observation, mean[i])
		if squaredDistance < minSquaredDistance {
			minSquaredDistance = squaredDistance
			indexOfCluster = i
		}
	}
	return indexOfCluster, math.Sqrt(minSquaredDistance)
}


func seed(data []ClusteredObservation, k int) []Observation {
	s := make([]Observation, k)
	rand.Seed(time.Now().Unix())
	for ii := 0; ii < k; ii++ {
		s[ii] = data[rand.Intn(len(data))].Observation
	}
	return s
}

// K-Means Algorithm 
func kmeans(data []ClusteredObservation,mean []Observation, threshold int) ([]Observation, error) {
	counter := 0
	for ii, jj := range data {
		closestCluster, _ := near(jj, mean)
		data[ii].ClusterNumber = closestCluster
	}

	mLen := make([]int, len(mean))

	for n := len(data[0].Observation); ; {
		for ii := range mean {
			mean[ii] = make(Observation, n)
			mLen[ii] = 0
		}
		for _, p := range data {
			mean[p.ClusterNumber].Add(p.Observation)
			mLen[p.ClusterNumber]++
		}
		for ii := range mean {
			mean[ii].Mul(1 / float64(mLen[ii]))
		}
		var changes int
		//numero pro =4

		numproc := 4
		end := make(chan bool)

		for i := 0; i < numproc; i++ {
			go func(id int) {
				for ii := len(data) % 4 * id; ii < len(data)/4*id; ii++ {
					if closestCluster, _ := near(data[ii], mean); closestCluster != data[ii].ClusterNumber {
						changes++
						data[ii].ClusterNumber = closestCluster
					}
				}

				end <- true
			}(i)
		}
		for i := 0; i < numproc; i++ {
			<-end
		}

		counter++
		if changes == 0 || counter > threshold {
			return mean, nil
		}
	}
	return mean, nil
}


func main() {
	const k = 3

	file, err := os.Open("number.txt")
	if err != nil {
		log.Fatal(err)
	}
	var contador = 0
	defer file.Close()
	scanner := bufio.NewScanner(file)
	observations := make([]ClusteredObservation, 55)
	for scanner.Scan() {

		var numbers = scanner.Text() + ""
		var n = strings.Split(numbers, " ")
		trimmed := strings.Trim(n[0], "[]")
		strings := strings.Split(trimmed, "\t")
		ints := make([]float64, len(strings))

		for i, s := range strings {
			var inter, _ = strconv.Atoi(s)
			ints[i] = float64(inter)
		}
		fmt.Printf("%#v\n", ints)
		data := make([]float64, 8)
		for ii := 0; ii < 8; ii++ {

			data[ii] = float64(ints[ii])

		}
		observations[contador].Observation = Observation(data)
		contador++
	}
	
	var inter = k
	seeds := seed(observations, inter)
	means, err = kmeans(observations, seeds, 10000)

	router := mux.NewRouter()
	router.HandleFunc("/foo",foo).Methods("POST")
	handler := cors.Default().Handler(router)
	http.ListenAndServe(":3000",handler)


}

func foo(w http.ResponseWriter, r *http.Request){
	decoder  := json.NewDecoder(r.Body)
	var data Data
	var pronostic Pronostic
	err := decoder.Decode(&data)
	if err != nil {
		fmt.Println("Data is missing")
		return 
	}
	var clusteredObservation ClusteredObservation 
	clusteredObservation.Observation = data.Values
	closestCluster, _ := near(clusteredObservation,means)
	if means[closestCluster][6] > 0.5 {
		pronostic.Infectado = true
	} else {
		pronostic.Infectado = false
	}
	pronostic.Riesgo = means[closestCluster][7]
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pronostic)

	diagnostico := setDiagnostico(data,pronostic)
	postBlockChain(diagnostico)
	fmt.Println(diagnostico)
}
func setDiagnostico(data Data,pronostic Pronostic)(Diagnostico){
	var diagnostico Diagnostico
	diagnostico.Edad = data.Values[0]
	diagnostico.Sexo = data.Values[1]
	diagnostico.Region = data.Values[2]
	diagnostico.Viaje = data.Values[3]
	diagnostico.InsuficienciaRespiratoria = data.Values[4]
	diagnostico.Neumonia = data.Values[5]
	diagnostico.Infectado = pronostic.Infectado
	diagnostico.Riesgo = pronostic.Riesgo
	return diagnostico
}

func postBlockChain(diagnostico Diagnostico){
	
	url := "http://localhost:8080"
	buf := new(bytes.Buffer)
	jsonValue, _ := json.Marshal(diagnostico)
	resp, _ := http.Post(url, "application/json", bytes.NewBuffer(jsonValue))
	json.NewEncoder(buf).Encode(diagnostico)
	fmt.Println(resp)

}
