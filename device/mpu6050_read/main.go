package main

import (
	"fmt"
	"machine"
	"math"
	"time"
	"tinygo.org/x/drivers/mpu6050"
)

type SensorValues struct {
	AX int32
	AY int32
	AZ int32
	GX int32
	GY int32
	GZ int32
}

var offset SensorValues = SensorValues{
	AX: 38233,
	AY: 20665,
	AZ: 32864,
	GX: 1244609,
	GY: 291075,
	GZ: 77860,
}

func readSensor(sensor *mpu6050.Device) *SensorValues {
	ax, ay, az := sensor.ReadAcceleration()
	gx, gy, gz := sensor.ReadRotation()
	return &SensorValues{
		AX: ax + offset.AX,
		AY: ay + offset.AY,
		AZ: az + offset.AZ,
		GX: gx + offset.GX,
		GY: gy + offset.GY,
		GZ: gz + offset.GZ,
	}
}

func calculateTiltAngles(data *SensorValues) (float32, float32, float32) {
	x := float64(data.AX)
	y := float64(data.AY)
	z := float64(data.AZ)
	return float32(math.Atan2(y, math.Sqrt(x*x+z*z))) * 180 / math.Pi, float32(math.Atan2(x, math.Sqrt(y*y+z*z)) * 180 / math.Pi), float32(math.Atan2(z, math.Sqrt(x*x+y*y)) * 180 / math.Pi)
}

func complementaryFilter(data *SensorValues, pitch *float64, roll *float64, delta_time float64, alpha float64) {
	x := float64(data.AX)
	y := float64(data.AY)
	z := float64(data.AZ)

	*pitch += x * delta_time
	*roll -= y * delta_time

	*pitch = alpha**pitch + (1-alpha)*math.Atan2(y, math.Sqrt(x*x+z*z))*180/math.Pi
	*roll = alpha**roll + (1-alpha)*math.Atan2(-x, math.Sqrt(y*y+z*z))*180/math.Pi
}

func main() {
	if err := machine.I2C0.Configure(machine.I2CConfig{}); err != nil {
		println(err)
		return
	}
	sensor := mpu6050.New(machine.I2C0)
	err := sensor.Configure()
	if err != nil {
		println(err)
		return
	}
	connected := sensor.Connected()
	if !connected {
		println("mpu6050 not connected")
		return
	}
	println("MPU6050 connected")

	data := &SensorValues{}
	var tmp float32
	var tiltX float32
	var tiltY float32
	var tiltZ float32
	var pitch float64
	var roll float64

	startMs := time.Now().UnixMilli()

	for {
		dt := float64(time.Now().UnixMilli()-startMs) / 1000
		// 温度を取得
		tmp = float32(machine.ReadTemperature()) / 1000

		// レジスタを指定して値を取得する場合（ここでは温度を取得）
		//data := make([]byte, 2)
		//if err := machine.I2C0.Tx(sensor.Address, []byte{mpu6050.TEMP_OUT_H}, data); err != nil {
		//	println(err)
		//	return
		//}
		//tmp := float32(int16((uint16(data[0])<<8)|uint16(data[1])))/340.0 + 36.53

		// 加速度と角速度を取得
		data = readSensor(&sensor)
		tiltX, tiltY, tiltZ = calculateTiltAngles(data)
		complementaryFilter(data, &pitch, &roll, dt, 0.98)

		fmt.Printf("Acceleration: %6.2f, %6.2f, %6.2f", float32(data.AX)/1000000, float32(data.AY)/1000000, float32(data.AZ)/1000000)
		fmt.Printf(" Rotation: %7.2f, %7.2f, %7.2f", float32(data.GX)/1000000, float32(data.GY)/1000000, float32(data.GZ)/1000000)
		fmt.Printf(" Temperature: %4.2f\n", tmp)
		fmt.Printf("Tilt: %6.2f, %6.2f %6.2f", tiltX, tiltY, tiltZ)
		fmt.Printf(" Pitch: %6.2f, Roll: %6.2f\n", pitch/1000000, roll/1000000)

		time.Sleep(time.Millisecond * 500)
	}
}
