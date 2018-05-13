"use strict";

var api_host, brick_host, wall_host;
var zoom_level, brick_id, base_image, shaderCanvas, gl, program;
if (window.location.hostname == 'localhost') {
    api_host='http://localhost:38000'
    brick_host='http://localhost:30080'
    wall_host='http://localhost:38001'
} else {
    api_host='http://wall-api.andrew-myhre.com'
    brick_host='http://brick.andrew-myhre.com'
    wall_host='http://wall.andrew-myhre.com'
}

zoom_level=1;
$(document).ready(function() {
    if (!window.location.hash) {
        alert("No brick id provided");
        return;
    }
    brick_id=window.location.hash.substring(1);
    //var img = new Image()
    //img.crossOrigin='Anonymous';
    //img.src = api_host+'/bricks/'+brick_id;
    var lc = null;
    var tools;
    var strokeWidths;
    var colors;

    var setCurrentByName;
    var findByName;

    // the only LC-specific thing we have to do
    var containerOne = document.getElementsByClassName('literally')[0];
    
    var showLC = function() {
        lc = LC.init(containerOne, {
            snapshot: JSON.parse(localStorage.getItem('drawing-'+brick_id)),
            defaultStrokeWidth: 10,
            strokeWidths: [10, 20, 50],
            secondaryColor: 'transparent',
            imageSize: {width: 1200, height: 675}
        });

        base_image = new Image();
        base_image.crossOrigin='Anonymous';
        base_image.src = api_host+'/bricks/'+brick_id;
        
        var b=$('.background')[0];
        b.width=lc.width;
        b.height=lc.height;
        window.demoLC = lc;

        shaderCanvas=document.getElementById('shaderCanvas');
        shaderCanvas.width=lc.width;
        shaderCanvas.height=lc.height;

        gl = shaderCanvas.getContext('webgl', {preserveDrawingBuffer: true});
        gl.viewport(0, 0, gl.drawingBufferWidth, gl.drawingBufferHeight);

        var buffer;
        buffer=gl.createBuffer();
        gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
        gl.bufferData(
            gl.ARRAY_BUFFER, 
            new Float32Array([
              -1.0, -1.0, 
               1.0, -1.0, 
              -1.0,  1.0, 
              -1.0,  1.0, 
               1.0, -1.0, 
               1.0,  1.0]), 
            gl.STATIC_DRAW
          );
        
        function drawImage(program, tex, texWidth, texHeight, dstX, dstY) {
            gl.bindTexture(gl.TEXTURE_2D, tex);
            
            // Tell WebGL to use our shader program pair
            gl.useProgram(program);
            
            // Setup the attributes to pull data from our buffers
            gl.bindBuffer(gl.ARRAY_BUFFER, positionBuffer);
            gl.enableVertexAttribArray(positionLocation);
            gl.vertexAttribPointer(positionLocation, 2, gl.FLOAT, false, 0, 0);
            gl.bindBuffer(gl.ARRAY_BUFFER, texcoordBuffer);
            gl.enableVertexAttribArray(texcoordLocation);
            gl.vertexAttribPointer(texcoordLocation, 2, gl.FLOAT, false, 0, 0);
            
            // this matirx will convert from pixels to clip space
            var matrix = m4.orthographic(0, gl.canvas.width, gl.canvas.height, 0, -1, 1);
            
            // this matrix will translate our quad to dstX, dstY
            matrix = m4.translate(matrix, dstX, dstY, 0);
            
            // this matrix will scale our 1 unit quad
            // from 1 unit to texWidth, texHeight units
            matrix = m4.scale(matrix, texWidth, texHeight, 1);
            
            // Set the matrix.
            gl.uniformMatrix4fv(matrixLocation, false, matrix);
            
            // Tell the shader to get the texture from texture unit 0
            gl.uniform1i(textureLocation, 0);
            
            // draw the quad (2 triangles, 6 vertices)
            gl.drawArrays(gl.TRIANGLES, 0, 6);
        }

        function verifyShaderCompiled(shader) {
            var compiled = gl.getShaderParameter(shader, gl.COMPILE_STATUS);
            if (!compiled)
            {
                var compilationLog = gl.getShaderInfoLog(shader);
                console.log(shader + ': ' + compilationLog);
                return false;
            }
            return true;
        }

        var vertexShader;
        var fragmentShader;

        var vertexShaderSource = `
        attribute vec4 a_position;
        attribute vec2 a_texCoord;
        
        uniform mat4 u_matrix;
        
        varying vec2 v_texCoord;
        
        void main() {
        gl_Position = u_matrix * a_position;
        v_texCoord = a_texCoord;
        }`;

        var fragmentShaderSource = `
        precision mediump float;
 
        varying vec2 v_texCoord;
        uniform sampler2D u_image;
        uniform vec2 u_textureSize;
        uniform float u_kernel_1[9];
        uniform float u_kernel_2[9];
        uniform float u_kernelWeight;
        
        void main() {
            vec2 onePixel = vec2(1.0, 1.0) / u_textureSize;
            vec4 colorSum =
                texture2D(u_image, v_texCoord + onePixel * vec2(-1, -1)) * u_kernel_1[0] +
                texture2D(u_image, v_texCoord + onePixel * vec2( 0, -1)) * u_kernel_1[1] +
                texture2D(u_image, v_texCoord + onePixel * vec2( 1, -1)) * u_kernel_1[2] +
                texture2D(u_image, v_texCoord + onePixel * vec2(-1,  0)) * u_kernel_1[3] +
                texture2D(u_image, v_texCoord + onePixel * vec2( 0,  0)) * u_kernel_1[4] +
                texture2D(u_image, v_texCoord + onePixel * vec2( 1,  0)) * u_kernel_1[5] +
                texture2D(u_image, v_texCoord + onePixel * vec2(-1,  1)) * u_kernel_1[6] +
                texture2D(u_image, v_texCoord + onePixel * vec2( 0,  1)) * u_kernel_1[7] +
                texture2D(u_image, v_texCoord + onePixel * vec2( 1,  1)) * u_kernel_1[8] ;

            colorSum *=
                texture2D(u_image, v_texCoord + onePixel * vec2(-1, -1)) * u_kernel_2[0] +
                texture2D(u_image, v_texCoord + onePixel * vec2( 0, -1)) * u_kernel_2[1] +
                texture2D(u_image, v_texCoord + onePixel * vec2( 1, -1)) * u_kernel_2[2] +
                texture2D(u_image, v_texCoord + onePixel * vec2(-1,  0)) * u_kernel_2[3] +
                texture2D(u_image, v_texCoord + onePixel * vec2( 0,  0)) * u_kernel_2[4] +
                texture2D(u_image, v_texCoord + onePixel * vec2( 1,  0)) * u_kernel_2[5] +
                texture2D(u_image, v_texCoord + onePixel * vec2(-1,  1)) * u_kernel_2[6] +
                texture2D(u_image, v_texCoord + onePixel * vec2( 0,  1)) * u_kernel_2[7] +
                texture2D(u_image, v_texCoord + onePixel * vec2( 1,  1)) * u_kernel_2[8] ;

            // Divide the sum by the weight but just use rgb
            // we'll set alpha to 1.0
            gl_FragColor = colorSum;
        }`;

        var kernelInUse='emboss';
        var kernels = {
            normal: [
              0, 0, 0,
              0, 1, 0,
              0, 0, 0
            ],
            gaussianBlur: [
              0.045, 0.122, 0.045,
              0.122, 0.332, 0.122,
              0.045, 0.122, 0.045
            ],
            gaussianBlur2: [
              1, 2, 1,
              2, 4, 2,
              1, 2, 1
            ],
            gaussianBlur3: [
              0, 1, 0,
              1, 1, 1,
              0, 1, 0
            ],
            unsharpen: [
              -1, -1, -1,
              -1,  9, -1,
              -1, -1, -1
            ],
            sharpness: [
               0,-1, 0,
              -1, 5,-1,
               0,-1, 0
            ],
            sharpen: [
               -1, -1, -1,
               -1, 16, -1,
               -1, -1, -1
            ],
            edgeDetect: [
               -0.125, -0.125, -0.125,
               -0.125,  1,     -0.125,
               -0.125, -0.125, -0.125
            ],
            edgeDetect2: [
               -1, -1, -1,
               -1,  8, -1,
               -1, -1, -1
            ],
            edgeDetect3: [
               -5, 0, 0,
                0, 0, 0,
                0, 0, 5
            ],
            edgeDetect4: [
               -1, -1, -1,
                0,  0,  0,
                1,  1,  1
            ],
            edgeDetect5: [
               -1, -1, -1,
                2,  2,  2,
               -1, -1, -1
            ],
            edgeDetect6: [
               -5, -5, -5,
               -5, 39, -5,
               -5, -5, -5
            ],
            sobelHorizontal: [
                1,  2,  1,
                0,  0,  0,
               -1, -2, -1
            ],
            sobelVertical: [
                1,  0, -1,
                2,  0, -2,
                1,  0, -1
            ],
            previtHorizontal: [
                1,  1,  1,
                0,  0,  0,
               -1, -1, -1
            ],
            previtVertical: [
                1,  0, -1,
                1,  0, -1,
                1,  0, -1
            ],
            boxBlur: [
                0.111, 0.111, 0.111,
                0.111, 0.111, 0.111,
                0.111, 0.111, 0.111
            ],
            triangleBlur: [
                0.0625, 0.125, 0.0625,
                0.125,  0.25,  0.125,
                0.0625, 0.125, 0.0625
            ],
            emboss: [
               -2, -1,  0,
               -1,  1,  1,
                0,  1,  2
            ]
          };

        vertexShader = gl.createShader(gl.VERTEX_SHADER);
        gl.shaderSource(vertexShader, vertexShaderSource);
        gl.compileShader(vertexShader);
        console.log ("verifying vertexShader");
        if (!verifyShaderCompiled(vertexShader)) {
            return;
        }
    
        fragmentShader = gl.createShader(gl.FRAGMENT_SHADER);
        gl.shaderSource(fragmentShader, fragmentShaderSource);
        gl.compileShader(fragmentShader);
        console.log ("verifying fragmentShader");
        if (!verifyShaderCompiled(fragmentShader)) {
        return;
        }

        program = gl.createProgram();
        gl.attachShader(program, vertexShader);
        gl.attachShader(program, fragmentShader);
        gl.linkProgram(program);	
        gl.getProgramInfoLog(program);
        gl.useProgram(program);

        // look up where the vertex data needs to go.
        var positionLocation = gl.getAttribLocation(program, "a_position");
        var texcoordLocation = gl.getAttribLocation(program, "a_texCoord");
        var textureSizeLocation = gl.getUniformLocation(program, "u_textureSize");
        var kernel1Location = gl.getUniformLocation(program, "u_kernel_1[0]");
        var kernel2Location = gl.getUniformLocation(program, "u_kernel_2[0]");
        var kernelWeightLocation = gl.getUniformLocation(program, "u_kernelWeight");

        function computeKernelWeight(kernel) {
            var weight = kernel.reduce(function(prev, curr) {
                return prev + curr;
            });
            return weight <= 0 ? 1 : weight;
        }

        // lookup uniforms
        var matrixLocation = gl.getUniformLocation(program, "u_matrix");
        var textureLocation = gl.getUniformLocation(program, "u_image");

        var positionBuffer = gl.createBuffer();
        gl.bindBuffer(gl.ARRAY_BUFFER, positionBuffer);
    
        // Put a unit quad in the buffer
        var positions = [
        0, 0,
        0, 1,
        1, 0,
        1, 0,
        0, 1,
        1, 1,
        ]
        gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(positions), gl.STATIC_DRAW);

        // Create a buffer for texture coords
        var texcoordBuffer = gl.createBuffer();
        gl.bindBuffer(gl.ARRAY_BUFFER, texcoordBuffer);

        // Put texcoords in the buffer
        var texcoords = [
            0, 0,
            0, 1,
            1, 0,
            1, 0,
            0, 1,
            1, 1,
        ]
        gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(texcoords), gl.STATIC_DRAW);
        
        
        base_image.onload = function()
        {
            var b=$('.background')[0];
            var ctx=b.getContext("2d");
            ctx.drawImage(base_image, 0, 0, 1200, 675);
        }

        function drawContextToGl(ctx) {
            var tex = gl.createTexture();
            gl.bindTexture(gl.TEXTURE_2D, tex);
            gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, ctx.getImageData(0,0,1200,675));

            // let's assume all images are not a power of 2
            gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
            gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
            gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
            gl.uniform2f(textureSizeLocation, 1200, 675);
            gl.uniform1fv(kernel1Location, kernels['emboss']);
            gl.uniform1fv(kernel2Location, kernels['gaussianBlur']);
            gl.uniform1f(kernelWeightLocation, computeKernelWeight(kernels[kernelInUse]));
    
            // Setup the attributes to pull data from our buffers
            gl.bindBuffer(gl.ARRAY_BUFFER, positionBuffer);
            // Tell the position attribute how to get data out of positionBuffer (ARRAY_BUFFER)
            var size = 2;          // 2 components per iteration
            var type = gl.FLOAT;   // the data is 32bit floats
            var normalize = false; // don't normalize the data
            var stride = 0;        // 0 = move forward size * sizeof(type) each iteration to get the next position
            var offset = 0;        // start at the beginning of the buffer
            gl.vertexAttribPointer(
                positionLocation, size, type, normalize, stride, offset)

            gl.enableVertexAttribArray(positionLocation);
            gl.vertexAttribPointer(positionLocation, 2, gl.FLOAT, false, 0, 0);
            
            gl.bindBuffer(gl.ARRAY_BUFFER, texcoordBuffer);
            // Tell the position attribute how to get data out of positionBuffer (ARRAY_BUFFER)
            var size = 2;          // 2 components per iteration
            var type = gl.FLOAT;   // the data is 32bit floats
            var normalize = false; // don't normalize the data
            var stride = 0;        // 0 = move forward size * sizeof(type) each iteration to get the next position
            var offset = 0;        // start at the beginning of the buffer
            gl.vertexAttribPointer(
                texcoordLocation, size, type, normalize, stride, offset)
            gl.enableVertexAttribArray(texcoordLocation);
            gl.vertexAttribPointer(texcoordLocation, 2, gl.FLOAT, false, 0, 0);
            
            var matrix = m4.orthographic(0, gl.canvas.width, gl.canvas.height, 0, -1, 1);
            matrix = m4.translate(matrix, 0, 0, 0);
            matrix = m4.scale(matrix, 1200, 675, 1);
            gl.uniformMatrix4fv(matrixLocation, false, matrix);
            gl.uniform1i(textureLocation, 0);
            gl.drawArrays(gl.TRIANGLES, 0, 6);
            
            gl.deleteTexture(tex);
        }

        function draw() {
            //webglUtils.resizeCanvasToDisplaySize(gl.canvas);
            var b=$('.background')[0];
            var ctx=b.getContext("2d");

            gl.clearColor(0.0, 0.0, 0.0, 1.0);
            gl.clear(gl.COLOR_BUFFER_BIT);

            // Tell WebGL how to convert from clip space to pixels
            gl.viewport(0, 0, gl.canvas.width, gl.canvas.height);
    
            // Tell WebGL to use our shader program pair
            gl.useProgram(program);
            
            drawContextToGl(ctx);
            //drawContextToGl(lc.ctx);
        }

        function render() {
            draw();
            requestAnimationFrame(render);
        }
        requestAnimationFrame(render);

        //render_gl();
    
        function render_gl() {
 
            window.requestAnimationFrame(render_gl, shaderCanvas);
          }
        
        function bounce() {
            $('.background')[0].getContext("2d").drawImage(lc.getImage({scaleDownRetina:true}),0,0);
            lc.clear();
        }

        var save = function() {
            localStorage.setItem('drawing-'+brick_id, JSON.stringify(lc.getSnapshot()));
        }

        lc.on('drawingChange', save);
        lc.on('shapeSave', bounce);
        lc.on('pan', save);
        lc.on('zoom', save);

        $("#open-image").click(function() {
            window.open(lc.getImage({
            scale: 1, margin: {top: 10, right: 10, bottom: 10, left: 10}
            }).toDataURL());
        });

        $("#change-size").click(function() {
            lc.setImageSize(null, 200);
        });

        $("#reset-size").click(function() {
            lc.setImageSize(null, null);
        });

        $("#clear-lc").click(function() {
            lc.clear();
        });
        $("#zoom-in").click(function() {
            lc.zoom(0.1)
            $('.background')[0].getContext("2d").scale(lc.scale,lc.scale)
        });
        $("#zoom-out").click(function() {
            lc.zoom(-0.1)
            $('.background')[0].getContext("2d").scale(lc.scale,lc.scale)
        });

        $("#publish-lc").click(function() {
            $('.background')[0].getContext("2d").drawImage(lc.getImage({scaleDownRetina:true}),0,0);
            $('.literally').hide();

            $.ajax({
                url: api_host+"/bricks/"+brick_id,
                type: "PUT",
                dataType: "json",
                data: JSON.stringify({
                    imagedata: $('.background')[0].toDataURL()
                }),
                processData: false,
                contentType: false,
                success: function(xhr, status) {
                    lc.clear();
                    save();
                    window.location.href = wall_host;
                },
                error: function(xhr, status) {
                },
                complete: function(xhr, status) {
                }
            });
        });

        // Set up our own tools...
        tools = [
            {
                name: 'tool-brush',
                el: document.getElementById('tool-brush'),
                tool: function() {
                    return new Brush(lc)
                }(LC.tools.ToolWithStroke)
            }
            /*,{
            name: 'pencil',
            el: document.getElementById('tool-pencil'),
            tool: new LC.tools.Pencil(lc)
            },{
            name: 'eraser',
            el: document.getElementById('tool-eraser'),
            tool: new LC.tools.Eraser(lc)
            },{
            name: 'text',
            el: document.getElementById('tool-text'),
            tool: new LC.tools.Text(lc)
            },{
            name: 'line',
            el: document.getElementById('tool-line'),
            tool: new LC.tools.Line(lc)
            },{
            name: 'arrow',
            el: document.getElementById('tool-arrow'),
            tool: function() {
                arrow = new LC.tools.Line(lc);
                arrow.hasEndArrow = true;
                return arrow;
            }()
            },{
            name: 'dashed',
            el: document.getElementById('tool-dashed'),
            tool: function() {
                dashed = new LC.tools.Line(lc);
                dashed.isDashed = true;
                return dashed;
            }()
            },{
            name: 'ellipse',
            el: document.getElementById('tool-ellipse'),
            tool: new LC.tools.Ellipse(lc)
            },{
            name: 'tool-rectangle',
            el: document.getElementById('tool-rectangle'),
            tool: new LC.tools.Rectangle(lc)
            },{
            name: 'tool-polygon',
            el: document.getElementById('tool-polygon'),
            tool: new LC.tools.Polygon(lc)
            },{
            name: 'tool-pan',
            el: document.getElementById('tool-pan'),
            tool: new LC.tools.Pan(lc)
            },{
            name: 'tool-select',
            el: document.getElementById('tool-select'),
            tool: new LC.tools.SelectShape(lc)
            }*/

        ];

        strokeWidths = [
            {
            name: 10,
            el: document.getElementById('sizeTool-1'),
            size: 10
            },{
            name: 20,
            el: document.getElementById('sizeTool-2'),
            size: 20
            },{
            name: 50,
            el: document.getElementById('sizeTool-3'),
            size: 50
            }
        ];

        colors = [
            {
            name: 'black',
            el: document.getElementById('colorTool-black'),
            color: '#000000'
            },{
            name: 'blue',
            el: document.getElementById('colorTool-blue'),
            color: '#0000ff'
            },{
            name: 'red',
            el: document.getElementById('colorTool-red'),
            color: '#ff0000'
            }
        ];

        setCurrentByName = function(ary, val) {
            ary.forEach(function(i) {
            $(i.el).toggleClass('current', (i.name == val));
            });
        };

        findByName = function(ary, val) {
            var vals;
            vals = ary.filter(function(v){
            return v.name == val;
            });
            if ( vals.length == 0 )
            return null;
            else
            return vals[0];
        };

        // Wire tools
        tools.forEach(function(t) {
            $(t.el).click(function() {
            var sw;

            lc.setTool(t.tool);
            setCurrentByName(tools, t.name);
            setCurrentByName(strokeWidths, t.tool.strokeWidth);
            $('#tools-sizes').toggleClass('disabled', (t.name == 'text'));
            });
        });
        setCurrentByName(tools, tools[0].name);
        lc.setTool(tools[0].tool);

        // Wire Stroke Widths
        // NOTE: This will not work until the stroke width PR is merged...
        strokeWidths.forEach(function(sw) {
            $(sw.el).click(function() {
            lc.trigger('setStrokeWidth', sw.size);
            setCurrentByName(strokeWidths, sw.name);
            })
        })
        setCurrentByName(strokeWidths, strokeWidths[0].name);

        // Wire Colors
        colors.forEach(function(clr) {
            $(clr.el).click(function() {
            lc.setColor('primary', clr.color)
            setCurrentByName(colors, clr.name);
            })
        })
        setCurrentByName(colors, colors[0].name);

    };

    $(document).ready(function() {
    // disable scrolling on touch devices so we can actually draw
    $(document).bind('touchmove', function(e) {
        if (e.target === document.documentElement) {
        return e.preventDefault();
        }
    });
    function toggleFullScreen() {
        if ((document.fullScreenElement && document.fullScreenElement !== null) ||    
         (!document.mozFullScreen && !document.webkitIsFullScreen)) {
          if (document.documentElement.requestFullScreen) {  
            document.documentElement.requestFullScreen();  
          } else if (document.documentElement.mozRequestFullScreen) {  
            document.documentElement.mozRequestFullScreen();  
          } else if (document.documentElement.webkitRequestFullScreen) {  
            document.documentElement.webkitRequestFullScreen(Element.ALLOW_KEYBOARD_INPUT);  
          }  
        } else {  
          if (document.cancelFullScreen) {  
            document.cancelFullScreen();  
          } else if (document.mozCancelFullScreen) {  
            document.mozCancelFullScreen();  
          } else if (document.webkitCancelFullScreen) {  
            document.webkitCancelFullScreen();  
          }  
        }  
      }
    $('#toggle-fullscreen').click(toggleFullScreen);
    showLC();
    });

    $('#hide-lc').click(function() {
    if (lc) {
        lc.teardown();
        lc = null;
    }
    });

    $('#show-lc').click(function() {
    if (!lc) { showLC(); }
    });

    $("#color-picker").spectrum({
        showButtons: false,
        change: function(color) {
            window.demoLC.setColor('primary', color.toRgbString())
            $('#color-picker').css('color',color.toRgbString())
        },
        move: function(color) {
            window.demoLC.setColor('primary', color.toRgbString())
            $('#color-picker').css('color',color.toRgbString())
        }
    });
});

$( window ).resize(function() {
});

function set_canvas_height() {
    aspect_ratio=$(document).width() / $(document).height();
    new_height=Math.floor($('.fs-container').width()/aspect_ratio);
    $('.fs-container').height(new_height+'px');
}
