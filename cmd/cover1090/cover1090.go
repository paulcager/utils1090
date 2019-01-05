package main

import (
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/golang/geo/r2"
)

const (
	dataFile = "/home/paul/locs.csv"
)

func main() {
	http.HandleFunc("/cover", coverHandler)
	http.HandleFunc("/", templateHandler)
	http.ListenAndServe(":8080", nil)
}

func coverHandler(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(dataFile)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Could not open %s: %s", dataFile, err)
		return
	}
	defer f.Close()

	//scanner := bufio.NewScanner(f)
	//for scanner.Scan() {
	//	fmt.Fscanf()
	//}

	var (
		points []r2.Point
		bounds r2.Rect
		x, y   float64
	)

	for i := 1; ; i++ {
		n, err := fmt.Fscanf(f, "%f,%f\n", &y, &x)
		if n == 2 {
			points = append(points, r2.Point{X: x, Y: y})
		}

		if err == io.EOF {
			break
		}

		//if err != nil {
		//	if err != io.EOF {
		//		w.WriteHeader(500)
		//		fmt.Fprintf(w, "Could not read %s at line %d: %s", dataFile, i, err)
		//		return
		//	}
		//	fmt.Printf("Read %d points\n", len(points))
		//	//bounds = r2.RectFromPoints(points...)
		//	break
		//}
	}

	bounds = r2.RectFromPoints(r2.Point{Y: 52, X: -4}, r2.Point{Y: 54, X: 0})
	i := create(1024, 1024, bounds, points)
	w.Header().Add("Content-Type", "image/png")
	png.Encode(w, i)
}

func create(width, height int, bounds r2.Rect, points []r2.Point) image.Image {
	i := image.NewNRGBA(image.Rect(0, 0, width, height))
	draw.Draw(i, i.Bounds(), &image.Uniform{color.Alpha{A: 128}}, image.ZP, draw.Src)
	xScale := float64(width) / (bounds.X.Hi - bounds.X.Lo)
	yScale := float64(height) / (bounds.Y.Hi - bounds.Y.Lo)

	count := 0
	for _, p := range points {
		if bounds.ContainsPoint(p) {
			i.Set(
				int((p.X-bounds.X.Lo)*xScale),
				int((p.Y-bounds.Y.Lo)*yScale),
				color.RGBA{0, 0, 255, 127},
			)
			//fmt.Println((p.X - bounds.X.Lo), int((p.X-bounds.X.Lo)*xScale), int((p.Y-bounds.Y.Lo)*yScale))
			count++
		}
	}

	fmt.Printf("Set %d points in (%f,%f)x(%f,%f)\n", count, bounds.X.Lo, bounds.Y.Lo, bounds.X.Hi, bounds.Y.Hi)
	fmt.Println("Scales", xScale, yScale)

	//for x := i.Bounds().Min.X; x <= i.Bounds().Max.X; x++ {
	//	for y := i.Bounds().Min.Y; y <= i.Bounds().Max.Y; y++ {
	//		i.Set(x, y, color.RGBA{0, uint8(y), 0, 255})
	//	}
	//
	//}
	return i
}

var templates = make(map[string]*template.Template)

func init() {
	dir := getEnv("TEMPLATES", "templates")
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		log.Printf("Could not scan template dir %q: %s", dir, err)
		os.Exit(2)
	}

	for _, f := range files {
		text, err := ioutil.ReadFile(f)
		if err != nil {
			log.Printf("Could not read file %q: %s", f, err)
			os.Exit(2)
		}
		name := f[len(dir)+1:]
		t := template.Must(template.New(name).Parse(string(text)))
		templates[name] = t
	}
}

func templateHandler(w http.ResponseWriter, r *http.Request) {
	values := map[string]interface{}{
		"APIKey": os.Getenv("APIKey"),
		"zoom":   getNumericEnv("ZOOM", 8),
		"lat":    getNumericEnv("LAT", 53),
		"lon":    getNumericEnv("LON", -2.25),
	}

	path := r.URL.Path
	if path == "/" || path == "" {
		path = "index"
	}

	t, ok := templates[path]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Println(path, "not found")
		return
	}

	err := t.Execute(w, values)
	if err != nil {
		fmt.Fprintln(w, err)
	}
}

func getEnv(key string, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getNumericEnv(key string, defaultValue float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return defaultValue
}
