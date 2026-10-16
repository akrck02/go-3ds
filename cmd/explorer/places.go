//go:build nintendo3ds

package main

import (
	"math"

	"github.com/akrck02/go-renderer/camera"
	"github.com/akrck02/go-renderer/graphics"
)

// place is somewhere on the island the camera can fly to.
type place struct {
	icon, name string
	target     graphics.Vec4
	distance   float64
	pitch      float64
}

// places are the island seen whole, its village on the terrace and its central peak.
var places = []place{
	{icon: "map", name: "Island", target: graphics.Vec4{0, 0, 0, 1}, distance: 450, pitch: 0.5},
	{icon: "house", name: "Village", target: graphics.Vec4{20, 0, 15, 1}, distance: 70, pitch: 0.32},
	{icon: "mountain", name: "Peak", target: graphics.Vec4{0, 30, 0, 1}, distance: 150, pitch: 0.7},
}

// flightTime is how long the camera takes to fly to a place.
const flightTime = 0.7

// flight eases the camera from where it is to a place.
type flight struct {
	active                 bool
	elapsed                float64
	fromTarget, toTarget   graphics.Vec4
	fromDistance, distance float64
	fromPitch, pitch       float64
}

func (journey *flight) start(orbit *camera.Orbit, to place) {
	*journey = flight{active: true,
		fromTarget: orbit.Target, toTarget: to.target,
		fromDistance: orbit.Distance, distance: to.distance,
		fromPitch: orbit.Pitch, pitch: to.pitch}
}

// step moves the camera along the flight: quickly at first, gently at the end.
func (flight *flight) step(orbit *camera.Orbit, seconds float64) {
	flight.elapsed += seconds
	progress := math.Min(flight.elapsed/flightTime, 1)
	eased := 1 - math.Pow(1-progress, 3)
	for axis := range 3 {
		orbit.Target[axis] = flight.fromTarget[axis] + (flight.toTarget[axis]-flight.fromTarget[axis])*float32(eased)
	}
	// Distance changes by ratio, so a flight closer looks as smooth as one away.
	orbit.Distance = flight.fromDistance * math.Pow(flight.distance/flight.fromDistance, eased)
	orbit.Pitch = flight.fromPitch + (flight.pitch-flight.fromPitch)*eased
	if progress >= 1 {
		flight.active = false
	}
}
