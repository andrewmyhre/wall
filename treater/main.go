package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"unsafe"

	"errors"
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

const (
	windowWidth  = 1600
	windowHeight = 900
)

var (
	triangle = []float32{
		-1.0, 1.0, 0, // top-left
		-1.0, -1.0, 0, // bottom-left
		1.0, -1.0, 0, // bottom-right
		-1.0, 1.0, 0,
		1.0, -1.0, 0,
		1.0, 1.0, 0,
	}
)

var emboss = []float32{
	-2, -1, 0,
	-1, 1, 1,
	0, 1, 2,
}

var gaussianBlur = []float32{
	0.045, 0.122, 0.045,
	0.122, 0.332, 0.122,
	0.045, 0.122, 0.045,
}

var vertexShaderSource = `
    #version 410
    in vec3 vp;
    void main() {
        gl_Position = vec4(vp, 1.0);
    }
` + "\x00"

var fragmentShaderSource = `
    #version 410
    out vec4 frag_colour;
    void main() {
        frag_colour = vec4(1, 1, 1, 1);
    }
` + "\x00"

type Texture struct {
	handle  uint32
	target  uint32 // same target as gl.BindTexture(<this param>, ...)
	texUnit uint32 // Texture unit that is currently bound to ex: gl.TEXTURE0
}

var errUnsupportedStride = errors.New("unsupported stride, only 32-bit colors supported")

var errTextureNotBound = errors.New("texture not bound")

func init() {
	// GLFW event handling must be run on the main OS thread
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to inifitialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(windowWidth, windowHeight, "basic shaders", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	// Initialize Glow (go function bindings)
	if err := gl.Init(); err != nil {
		panic(err)
	}

	window.SetKeyCallback(keyCallback)

	err = programLoop(window)
	if err != nil {
		log.Fatal(err)
	}
}

func programLoop(window *glfw.Window) error {
	vertices := []float32{
		// top left
		-1.0, 1.0, 0.0, // position
		1.0, 0.0, 0.0, // Color
		0.0, 1.0, // texture coordinates

		// top right
		1.0, 1.0, 0.0,
		0.0, 1.0, 0.0,
		1.0, 1.0,

		// bottom right
		1.0, -1.0, 0.0,
		0.0, 0.0, 1.0,
		1.0, 0.0,

		// bottom left
		-1.0, -1.0, 0.0,
		1.0, 1.0, 1.0,
		0.0, 0.0,
	}

	indices := []uint32{
		// rectangle
		0, 1, 2, // top triangle
		0, 2, 3, // bottom triangle
	}

	vertShader, err := NewShaderFromFile("shaders/paint.vert", gl.VERTEX_SHADER)
	if err != nil {
		return err
	}

	fragShader, err := NewShaderFromFile("shaders/paint.frag", gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	shaderProgram, err := NewProgram(vertShader, fragShader)
	if err != nil {
		return err
	}
	defer shaderProgram.Delete()

	VAO := createVAO(vertices, indices)
	texture0, err := NewTextureFromFile("img.jpg", gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}

	glfw.PollEvents()

	// background color
	gl.ClearColor(0.2, 0.5, 0.5, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	// draw vertices
	shaderProgram.Use()

	// set texture0 to uniform0 in the fragment shader
	texture0.Bind(gl.TEXTURE0)
	texture0.SetUniform(shaderProgram.GetUniformLocation("u_image"))

	gl.Uniform2f(shaderProgram.GetUniformLocation("u_textureSize"), windowWidth, windowHeight)
	gl.Uniform1fv(shaderProgram.GetUniformLocation("u_kernel_1"), 9, &emboss[0])
	gl.Uniform1fv(shaderProgram.GetUniformLocation("u_kernel_2"), 9, &gaussianBlur[0])
	gl.Uniform1f(shaderProgram.GetUniformLocation("u_kernelWeight"), 1.0)

	gl.BindVertexArray(VAO)
	gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, unsafe.Pointer(nil))
	gl.BindVertexArray(0)

	texture0.UnBind()

	screenshot := image.NewRGBA(image.Rect(0, 0, int(windowWidth), int(windowHeight)))
	gl.ReadPixels(0, 0, windowWidth, windowHeight, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(screenshot.Pix))

	fo, err := os.Create("img_treated.png")
	if err != nil {
		panic(err)
	}
	defer fo.Close()

	err = png.Encode(fo, screenshot)

	if err != nil {
		panic(err)
	}
	// end of draw loop

	// swap in the rendered buffer
	window.SwapBuffers()

	return nil
}

// initGlfw initializes glfw and returns a Window to use.
func initGlfw() *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4) // OR 2
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(windowWidth, windowHeight, "treater", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	return window
}

func Render(vao uint32, window *glfw.Window, program uint32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	gl.BindVertexArray(vao)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(triangle)/3))

	glfw.PollEvents()
	window.SwapBuffers()
}

func createVAO(vertices []float32, indices []uint32) uint32 {

	var VAO uint32
	gl.GenVertexArrays(1, &VAO)

	var VBO uint32
	gl.GenBuffers(1, &VBO)

	var EBO uint32
	gl.GenBuffers(1, &EBO)

	// Bind the Vertex Array Object first, then bind and set vertex buffer(s) and attribute pointers()
	gl.BindVertexArray(VAO)

	// copy vertices data into VBO (it needs to be bound first)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// copy indices into element buffer
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, EBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// size of one whole vertex (sum of attrib sizes)
	var stride int32 = 3*4 + 3*4 + 2*4
	var offset int = 0

	// position
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(offset))
	gl.EnableVertexAttribArray(0)
	offset += 3 * 4

	// color
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride, gl.PtrOffset(offset))
	gl.EnableVertexAttribArray(1)
	offset += 3 * 4

	// texture position
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, stride, gl.PtrOffset(offset))
	gl.EnableVertexAttribArray(2)
	offset += 2 * 4

	// unbind the VAO (safe practice so we don't accidentally (mis)configure it later)
	gl.BindVertexArray(0)

	return VAO
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func (tex *Texture) Bind(texUnit uint32) {
	gl.ActiveTexture(texUnit)
	gl.BindTexture(tex.target, tex.handle)
	tex.texUnit = texUnit
}

func (tex *Texture) UnBind() {
	tex.texUnit = 0
	gl.BindTexture(tex.target, 0)
}

func (tex *Texture) SetUniform(uniformLoc int32) error {
	if tex.texUnit == 0 {
		return errTextureNotBound
	}
	gl.Uniform1i(uniformLoc, int32(tex.texUnit-gl.TEXTURE0))
	return nil
}

func NewTextureFromFile(file string, wrapR, wrapS int32) (*Texture, error) {
	imgFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer imgFile.Close()

	// Decode detexts the type of image as long as its image/<type> is imported
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, err
	}
	return NewTexture(img, wrapR, wrapS)
}

func NewTexture(img image.Image, wrapR, wrapS int32) (*Texture, error) {
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Pt(0, 0), draw.Src)
	if rgba.Stride != rgba.Rect.Size().X*4 { // TODO-cs: why?
		return nil, errUnsupportedStride
	}

	var handle uint32
	gl.GenTextures(1, &handle)

	target := uint32(gl.TEXTURE_2D)
	internalFmt := int32(gl.SRGB_ALPHA)
	format := uint32(gl.RGBA)
	width := int32(rgba.Rect.Size().X)
	height := int32(rgba.Rect.Size().Y)
	pixType := uint32(gl.UNSIGNED_BYTE)
	dataPtr := gl.Ptr(rgba.Pix)

	texture := Texture{
		handle: handle,
		target: target,
	}

	texture.Bind(gl.TEXTURE0)
	defer texture.UnBind()

	// set the texture wrapping/filtering options (applies to current bound texture obj)
	// TODO-cs
	gl.TexParameteri(texture.target, gl.TEXTURE_WRAP_R, wrapR)
	gl.TexParameteri(texture.target, gl.TEXTURE_WRAP_S, wrapS)
	gl.TexParameteri(texture.target, gl.TEXTURE_MIN_FILTER, gl.LINEAR) // minification filter
	gl.TexParameteri(texture.target, gl.TEXTURE_MAG_FILTER, gl.LINEAR) // magnification filter

	gl.TexImage2D(target, 0, internalFmt, width, height, 0, format, pixType, dataPtr)

	gl.GenerateMipmap(texture.handle)

	return &texture, nil
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action,
	mods glfw.ModifierKey) {

	// When a user presses the escape key, we set the WindowShouldClose property to true,
	// which closes the application
	if key == glfw.KeyEscape && action == glfw.Press {
		window.SetShouldClose(true)
	}
}
