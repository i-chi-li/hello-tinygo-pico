package main

import (
	"fmt"
	"machine"
	"math"
	"time"
	"tinygo.org/x/drivers/mpu6050"
)

const BUFFER_SIZE = 1000
const ACCEL_DEAD_ZONE = 1000
const GYRO_DEAD_ZONE = 1000

type SensorValues struct {
	AX int32
	AY int32
	AZ int32
	GX int32
	GY int32
	GZ int32
}

var offset SensorValues = SensorValues{
	AX: 0,
	AY: 0,
	AZ: 0,
	GX: 0,
	GY: 0,
	GZ: 0,
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

func meanSensors(sensor *mpu6050.Device, mean *SensorValues) {
	buff := SensorValues{}

	for i := 0; i < BUFFER_SIZE+101; i++ {
		data := readSensor(sensor)
		if i > 100 && i <= BUFFER_SIZE+100 {
			buff.AX += data.AX
			buff.AY += data.AY
			buff.AZ += data.AZ
			buff.GX += data.GX
			buff.GY += data.GY
			buff.GZ += data.GZ
		}
		if i == BUFFER_SIZE+100 {
			mean.AX = buff.AX / BUFFER_SIZE
			mean.AY = buff.AY / BUFFER_SIZE
			mean.AZ = buff.AZ / BUFFER_SIZE
			mean.GX = buff.GX / BUFFER_SIZE
			mean.GY = buff.GY / BUFFER_SIZE
			mean.GZ = buff.GZ / BUFFER_SIZE
		}
		time.Sleep(time.Millisecond)
	}
}

func calibration(sensor *mpu6050.Device, mean *SensorValues) {
	tmp := SensorValues{
		AX: -mean.AX / 8,
		AY: -mean.AY / 8,
		AZ: (1000000 - mean.AZ) / 8,
		GX: -mean.GX / 4,
		GY: -mean.GY / 4,
		GZ: -mean.GZ / 4,
	}

	for {
		ready := 0
		offset.AX = tmp.AX
		offset.AY = tmp.AY
		offset.AZ = tmp.AZ
		offset.GX = tmp.GX
		offset.GY = tmp.GY
		offset.GZ = tmp.GZ

		meanSensors(sensor, mean)
		fmt.Printf("Offsets: %4d, %4d, %4d, %4d, %4d, %4d ", offset.AX, offset.AY, offset.AZ, offset.GX, offset.GY, offset.GZ)
		fmt.Printf("Mean: %4d, %4d, %4d, %4d, %4d, %4d ", mean.AX, mean.AY, mean.AZ, mean.GX, mean.GY, mean.GZ)
		fmt.Printf("Tmp: %4d, %4d, %4d, %4d, %4d, %4d\n", tmp.AX, tmp.AY, tmp.AZ, tmp.GX, tmp.GY, tmp.GZ)

		if math.Abs(float64(mean.AX)) <= ACCEL_DEAD_ZONE {
			ready++
		} else {
			tmp.AX -= mean.AX / 8
		}
		if math.Abs(float64(mean.AY)) <= ACCEL_DEAD_ZONE {
			ready++
		} else {
			tmp.AY -= mean.AY / 8
		}
		if math.Abs(float64(1000000-mean.AZ)) <= ACCEL_DEAD_ZONE {
			ready++
		} else {
			tmp.AZ += (1000000 - mean.AZ) / 8
		}
		if math.Abs(float64(mean.GX)) <= GYRO_DEAD_ZONE {
			ready++
		} else {
			tmp.GX -= mean.GX / 4
		}
		if math.Abs(float64(mean.GY)) <= GYRO_DEAD_ZONE {
			ready++
		} else {
			tmp.GY -= mean.GY / 4
		}
		if math.Abs(float64(mean.GZ)) <= GYRO_DEAD_ZONE {
			ready++
		} else {
			tmp.GZ -= mean.GZ / 4
		}

		if ready == 6 {
			break
		}
	}
}

func main() {
	if err := machine.I2C0.Configure(machine.I2CConfig{}); err != nil {
		fmt.Println(err)
		return
	}
	sensor := mpu6050.New(machine.I2C0)
	err := sensor.Configure()
	if err != nil {
		fmt.Println(err)
		return
	}
	connected := sensor.Connected()
	if !connected {
		fmt.Println("mpu6050 not connected")
		return
	}
	fmt.Println("MPU6050 connected")

	for machine.Serial.Buffered() > 0 {
		machine.Serial.ReadByte()
	}
	for machine.Serial.Buffered() == 0 {
		println("Send any charactoer to start sketch.")
		time.Sleep(time.Second * 2)
	}
	for machine.Serial.Buffered() > 0 {
		machine.Serial.ReadByte()
	}
	println("MPU6050 Calibration Sketch")
	time.Sleep(time.Second * 2)
	println("Your MPU6050 should be placed in horizontal position, with package letters facing up. \n" +
		"Don't touch it until you see a finish message.")
	time.Sleep(time.Second * 3)

	state := 0
	mean := SensorValues{}

	for {
		if state == 0 {
			println("Reading sensors for first time...")
			meanSensors(&sensor, &mean)
			state++
			time.Sleep(time.Second)
		}

		if state == 1 {
			println("Calculating offsets...")
			calibration(&sensor, &mean)
			state++
			time.Sleep(time.Second)
		}

		if state == 2 {
			meanSensors(&sensor, &mean)
			println("FINISHED!")
			println("Sensor readings with offsets:")
			println(mean.AX, mean.AY, mean.AZ, mean.GX, mean.GY, mean.GZ)
			println("Your offsets:")
			println(offset.AX, offset.AY, offset.AZ, offset.GX, offset.GY, offset.GZ)

			println("Data is printed as: accelX accelY accelZ gyroX gyroY gyroZ")
			println("Check that your sensor readings are close to 0 0 1000000 0 0 0")
			println("If calibration was successful write down your offsets so you can set them in your projects using somethings similar")
			return
		}
	}
}
