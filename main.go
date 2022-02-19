package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	frame_width   = 960
	frame_height  = 960
	Width_count   = 12
	Height_count  = 16
	Height        = 60
	Width         = 80
	default_dir   = "img/"
	res_dir       = "res/"
	final_res_dir = "final_res/"
	letter_dir    = "letters/"
	null          = "null"
	temp_path     = "temp/template.png"
	null_img_path = "temp/null.png"
)

var (
	temp     image.Image
	null_img image.Image
)

type Changeable interface {
	Set(x, y int, c color.Color)
}

type cell struct {
	value               string
	not_accepted_values map[string]int
}

func new_cell(value string, not_accepted_values []string) cell {
	return cell{
		value:               value,
		not_accepted_values: make(map[string]int),
	}
}

func roll_rand(max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max)
}

func def_max_flag_count(flag_count int, grid_size int) (int, []int) {
	max := make([]int, flag_count)
	for i := range max {
		max[i] = 0
	}
	return int(grid_size/flag_count) + 1, max
}
func print_grid(grid [][]cell) {
	for i := 0; i < Height_count; i++ {
		for j := 0; j < Width_count; j++ {
			fmt.Print(grid[i][j].value, " ")
		}
		fmt.Println()
	}
}
func create_rand_grid(grid [][]cell, flag []string) {
	flags_max, stat := def_max_flag_count(len(flag), Width_count*Height_count)
	for i := 0; i < Height_count; i++ {
		for j := 0; j < Width_count; j++ {
			var x int
			possible_combo := false
			for it := range flag {
				if stat[it] < flags_max {
					if val, ok := grid[i][j].not_accepted_values[flag[it]]; ok {
						if val == -1 {
							possible_combo = true
						}
					} else {
						grid[i][j].not_accepted_values[flag[it]] = -1
						possible_combo = true
					}
				}
			}

			//ac accepted value
			//ac accepted count
			for av, ac := true, false; !(av && ac); {
				av, ac = true, false
				x = roll_rand(len(flag))
				if possible_combo {
					if val, ok := grid[i][j].not_accepted_values[flag[x]]; ok {
						if val != -1 {
							av = false
						}
					}
					if stat[x] < flags_max {
						ac = true

					}
				} else {
					break
				}

			}
			if possible_combo {
				grid[i][j].value = flag[x]
				stat[x]++
				if i > 0 {
					grid[i-1][j].not_accepted_values[flag[x]] = 1
				}
				if i < Height_count-1 {
					grid[i+1][j].not_accepted_values[flag[x]] = 1
				}
				if j > 0 {
					grid[i][j-1].not_accepted_values[flag[x]] = 1
				}
				if j < Width_count-1 {
					grid[i][j+1].not_accepted_values[flag[x]] = 1
				}
			} else {
				grid[i][j].value = null
			}

		}

	}
}
func save_img(path string, img image.Image) error {
	fmt.Printf("saving image : %s.png\n", path)
	f, _ := os.Create(path + ".png")
	err := png.Encode(f, img)
	defer f.Close()
	return err
}
func create_rand_bg(grid [][]cell, m map[string]image.Image, id int, key byte) {
	for j := 0; j < Height_count; j++ {
		for i := 0; i < Width_count; i++ {
			for y := 0; y < Height; y++ {
				for x := 0; x < Width; x++ {
					if grid[j][i].value == null {
						temp.(Changeable).Set(i*Width+x, j*Height+y, null_img.At(x, y))
					} else {
						temp.(Changeable).Set(i*Width+x, j*Height+y, m[grid[j][i].value].At(x, y))
					}

				}
			}
		}
	}
	save_img(res_dir+string(key)+strconv.Itoa(id), temp)
}
func walk_dir(path string) map[string]image.Image {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	m := make(map[string]image.Image)

	for _, file := range files {
		img := decode_png(path, file.Name())
		m[file.Name()[:len(file.Name())-len(filepath.Ext(file.Name()))]] = img
	}
	return m
}
func reset_grid(grid [][]cell) {
	for i := 0; i < Height_count; i++ {
		grid[i] = make([]cell, Width_count)
		for j := 0; j < Width_count; j++ {
			grid[i][j] = new_cell("", []string{})
		}
	}
}

func fit_letter(letters map[string]image.Image, m map[string]image.Image) {
	for lk := range letters {
		for ik := range m {
			if strings.ToLower(lk)[0] == strings.ToLower(ik)[0] {
				letter_copy := decode_png(letter_dir, lk+".png")
				for x := 0; x <= frame_width; x++ {
					for y := 0; y <= frame_height; y++ {
						r, _, _, _ := letter_copy.At(x, y).RGBA()
						if r != 0 {

							letter_copy.(Changeable).Set(x, y, m[ik].At(x, y))
						}

					}
				}
				save_img(final_res_dir+lk+ik, letter_copy)
			}

		}

	}

}
func decode_png(dir string, name string) image.Image {

	imgfile, err := os.Open(dir + name)
	if err != nil {
		panic(err.Error())
	}
	defer imgfile.Close()
	img, err := png.Decode(imgfile)
	if err != nil {
		panic(err.Error())
	}
	return img
}
func sort(m map[string]image.Image) map[byte][]string {
	res := make(map[byte][]string)
	for k := range m {
		short_k := strings.ToLower(k)
		if _, ok := res[short_k[0]]; ok {
			res[short_k[0]] = append(res[short_k[0]], k)
		} else {
			res[short_k[0]] = []string{k}
		}
	}
	return res
}
func main() {
	temp = decode_png(temp_path, "")
	null_img = decode_png(null_img_path, "")
	m := walk_dir(default_dir)
	sorted_flags := sort(m)
	grid := make([][]cell, Height_count)
	reset_grid(grid)

	for key, val := range sorted_flags {
		for k := 0; k < len(val); k++ {
			reset_grid(grid)
			create_rand_grid(grid, val)
			create_rand_bg(grid, m, k, key)
		}
	}

	letters := walk_dir(letter_dir)
	res := walk_dir(res_dir)
	fit_letter(letters, res)

}
