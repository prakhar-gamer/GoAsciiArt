package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
	"golang.org/x/term"
)

func main() {
	//sets the directory for the location of the files
	homeBase := createMainDir()

	fmt.Print("Please enter file directory for where the mp3 is located: ")
	var fileDirectory string
	_, err := fmt.Scanln(&fileDirectory)
	if err != nil {
		fmt.Println("Bad directory")
		return
	}

	fmt.Print("Please enter File Name of the mp3 (include the .mp3 at the end): ")
	var fileName string
	_, errs := fmt.Scanln(&fileName)
	if errs != nil {
		fmt.Println("Bad name")
		return
	}
	imageGenerator(fileName, fileDirectory, homeBase)
}

// preconditions: has the title of the song, the location of the song the mp3, and a directory to make the file to convert the mp3
// This is going to check for the image being a mp3, rip the image, and run the acsii art convertyer
func imageGenerator(title string, loco string, home string) {
	fileLoco := fileLocator(title, loco)

	//checks the extention is a mp3
	if filepath.Ext(fileLoco) == ".mp3" {
		coverlocation := extractCover(fileLoco, home, title)

		//gets the width of the terminal
		//REMINDER: Don't add the height too cause then it will stretch the art, better to make the art keep its aspect ratio ygwim no cappy
		width, _, err := term.GetSize(int(os.Stdin.Fd()))
		if err != nil {
			width = 80 // safety width, so program don't crash
		}

		//using the function acsii art with the location of the cover, and the cover location var I made sends the rest of the job to that function
		imgArt, err := acsii_art(coverlocation, width)
		if err != nil {
			fmt.Print(err)
			return
		}

		//prints the art
		fmt.Println("\n" + imgArt)
	} else {
		//to make sure program dont exploady.
		fmt.Println("This isn't a mp3")
	}
}

// gets a cover from a mp3
func extractCover(fileLoco string, home string, title string) string {
	//"opens" the file, but like its like a clientless file opener
	songFile, err := os.Open(fileLoco)
	if err != nil {
		fmt.Println(err)
	}
	//it makes sure the songfile closes no matter what
	defer songFile.Close()

	//this uses the tag libary, and gets the info from the file
	songInfo, err := tag.ReadFrom(songFile)
	if err != nil {
		fmt.Println(err)
	}

	//gets the image from the song, and if it doesn't have a cover then it just prints that it has no picture and returns to the main function with the err function
	songPic := songInfo.Picture()
	if songPic == nil {
		fmt.Print("no picture")
		return "err"
	}

	// cover extention handling
	ext := songPic.Ext
	//if the file ends up having no extention
	if ext == "" {
		//splits the name name into parts, and checks for jpeg if it is then turns it into a jpg
		parts := strings.Split(songPic.MIMEType, "/")
		if len(parts) == 2 && parts[0] == "image" {
			ext = parts[1]
			if ext == "png" {
				ext = "jpg"
			}
		} else {
			ext = "jpg"
		}
	}

	// title for the Cover
	//list of char not allowed in the File explorers
	invalidChars := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}
	//isolates the file name from the extenstion
	coverFileName := strings.TrimSuffix(title, ".mp3")
	//goes through the entire list of char for the title, too make sure non are those
	for _, char := range invalidChars {
		coverFileName = strings.ReplaceAll(coverFileName, char, "_")
	}
	coverFileName = strings.TrimSpace(coverFileName)
	if coverFileName == "" {
		coverFileName = "untitled"
	}

	//connects the new coverFileName and entenstion to just become the new file. Allows to do multiple different images, as it doesn't let the same image exist twice i think
	coverName := fmt.Sprintf("%s.%s", coverFileName, ext)
	outputPath := filepath.Join(home, coverName)

	//creates the file with the custom file name, at the chosen path
	err = os.WriteFile(outputPath, songPic.Data, 0644)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print("Done! Check folder")
	return outputPath

}

//creates the directory for the music cli for covers

func createMainDir() string {
	fmt.Println("Where do you want to create a file for the cover image location?")
	var locationDir string
	//gets what the user typed and then stores it to a variable
	_, filepatherr := fmt.Scanln(&locationDir)
	if filepatherr != nil {
		fmt.Print(filepatherr)
	}

	//creates the folder at the chose path
	folderName := "musicli"
	createPath := filepath.Join(locationDir, folderName)
	//just checks if their was any err in creating the path, and then prints if their was and returns to the main functions
	folderCreateError := os.MkdirAll(createPath, 0755)
	if folderCreateError != nil {
		fmt.Println(folderCreateError)
		return ""
	}
	fmt.Println("It worked!")
	return createPath
}

// finds the file with only the name
func fileLocator(title string, loco string) string {
	//it creates a array that reads through the file path it gave u, and adds every file to the array
	entries, err := os.ReadDir(loco)
	if err != nil {
		return "Cant Reach location"
	}

	//loops through all the entires waiting for the exact name so
	found := false
	for _, entry := range entries {
		if !entry.IsDir() {
			fileName := filepath.Base(entry.Name())
			if fileName == title {
				found = true
				locoFilePath := filepath.Join(loco, fileName)
				return locoFilePath
			}
		}
	}
	if !found {
		return "NOT THEIR DUMMY"
	}
	return ""
}

// precondition: has the locations and the terminal width
// post condition: makes the ascii art and returns as a string with color in the terminal
// creates the art from the cover
func acsii_art(loco string, termW int) (string, error) {
	//makes sure the image is accessiable and won't cause a problemo
	startImage, err := os.Open(loco)
	if err != nil {
		return "", err
	}
	defer startImage.Close()

	//converts the image into data packet thingys (bits according to google)
	img, _, err := image.Decode(startImage)
	if err != nil {
		return "", err
	}

	//gets the the image size
	bounds := img.Bounds()
	imgW, imgH := bounds.Max.X, bounds.Max.Y

	// gets the aspect ratio, then uses very mathy math (takes te term W times aspect ratio times .4 (cuz it looks nicer)) too then get term H
	aspectRatio := float64(imgH) / float64(imgW)
	termH := int(float64(termW) * aspectRatio * 0.4)

	//basically a brush from density to like empty + the var we return and actually is the color
	var sb strings.Builder
	ramp := "@#W$9876543210?!abc;:+=-,._ "
	rampLen := float64(len(ramp) - 1)

	// degrades the picture to make the size of each pixel
	stepX := float64(imgW) / float64(termW)
	stepY := float64(imgH) / float64(termH)

	//repeats for every row of "bits" of image
	for y := 0; y < termH; y++ {
		for x := 0; x < termW; x++ {
			imgX := int(float64(x) * stepX)
			imgY := int(float64(y) * stepY)

			// safety bound checks
			if imgX >= imgW {
				imgX = imgW - 1
			}
			if imgY >= imgH {
				imgY = imgH - 1
			}

			//gets the rgb values for the bit its on for the color
			pixel := img.At(imgX, imgY)
			r, g, b, _ := pixel.RGBA()
			rv, gv, bv := uint8(r>>8), uint8(g>>8), uint8(b>>8)

			//converts the pixel into a grayscale cover then using then picks are charecvter that best represents
			grayColor := color.GrayModel.Convert(pixel).(color.Gray)
			intensity := float64(grayColor.Y) / 255.0
			char := ramp[int(intensity*rampLen)]

			// builds the string with the pixel, and color
			sb.WriteString(fmt.Sprintf("\033[38;2;%d;%d;%dm%c\033[0m", rv, gv, bv, char))
		}
		sb.WriteString("\n")
	}
	return sb.String(), nil
}
