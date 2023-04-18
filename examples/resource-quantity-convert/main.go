package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/api/resource"
)

func main() {
	// 1. 带单位的字符串转换为 resource.Quantity
	{
		memorySize := resource.MustParse("5Gi")
		fmt.Printf("memorySize = %v (%v)\n", memorySize.Value(), memorySize.Format)

		diskSize := resource.MustParse("5G")
		fmt.Printf("diskSize = %v (%v)\n", diskSize.Value(), diskSize.Format)

		cores := resource.MustParse("5300m")
		fmt.Printf("milliCores = %v (%v)\n", cores.MilliValue(), cores.Format)

		cores2 := resource.MustParse("5.4")
		fmt.Printf("milliCores = %v (%v)\n", cores2.MilliValue(), cores2.Format)

		// 输出结果:
		// memorySize = 5368709120 (BinarySI)
		// diskSize = 5000000000 (DecimalSI)
		// milliCores = 5300 (DecimalSI)
		// milliCores = 5400 (DecimalSI)
	}

	// 2. 数值转换为 resource.Quantity
	{
		memorySize := resource.NewQuantity(5*1024*1024*1024, resource.BinarySI)
		fmt.Printf("memorySize = %v\n", memorySize) //

		diskSize := resource.NewQuantity(5*1000*1000*1000, resource.DecimalSI)
		fmt.Printf("diskSize = %v\n", diskSize)

		cores := resource.NewMilliQuantity(5300, resource.DecimalSI)
		fmt.Printf("cores = %v\n", cores)

		// 输出结果:
		// memorySize = 5Gi
		// diskSize = 5G
		// cores = 5300m
	}

	// 3. resource.Quantity 参数解析
	{
		fmt.Println("=> Parse resource.Quantity from command line arguments...")
		q := resource.QuantityValue{
			Quantity: resource.MustParse("1Mi"),
		}
		fs := pflag.FlagSet{}
		fs.SetOutput(os.Stdout)
		fs.Var(&q, "mem", "sets amount of memory")
		fs.PrintDefaults()

		fs.Parse(os.Args[1:])
		fmt.Printf("q value: %v\n", q.String())

		// go run main.go --mem 5Gi
		// 输出结果:
		// --mem quantity   sets amount of memory (default 1Mi)
		// q value: 5Gi
	}
}
