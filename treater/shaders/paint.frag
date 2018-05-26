#version 410 core

in vec3 ourColor;
in vec2 v_texCoord;
uniform sampler2D u_image;
uniform vec2 u_textureSize;
uniform float u_kernel_1[9];
uniform float u_kernel_2[9];
uniform float u_kernelWeight;
out vec4 color;

void main() {
    vec2 onePixel = vec2(1.0, 1.0) / u_textureSize;
    vec4 colorSum =
        texture(u_image, v_texCoord + onePixel * vec2(-1, -1)) * u_kernel_1[0] +
        texture(u_image, v_texCoord + onePixel * vec2( 0, -1)) * u_kernel_1[1] +
        texture(u_image, v_texCoord + onePixel * vec2( 1, -1)) * u_kernel_1[2] +
        texture(u_image, v_texCoord + onePixel * vec2(-1,  0)) * u_kernel_1[3] +
        texture(u_image, v_texCoord + onePixel * vec2( 0,  0)) * u_kernel_1[4] +
        texture(u_image, v_texCoord + onePixel * vec2( 1,  0)) * u_kernel_1[5] +
        texture(u_image, v_texCoord + onePixel * vec2(-1,  1)) * u_kernel_1[6] +
        texture(u_image, v_texCoord + onePixel * vec2( 0,  1)) * u_kernel_1[7] +
        texture(u_image, v_texCoord + onePixel * vec2( 1,  1)) * u_kernel_1[8] ;

    colorSum *=
        texture(u_image, v_texCoord + onePixel * vec2(-1, -1)) * u_kernel_2[0] +
        texture(u_image, v_texCoord + onePixel * vec2( 0, -1)) * u_kernel_2[1] +
        texture(u_image, v_texCoord + onePixel * vec2( 1, -1)) * u_kernel_2[2] +
        texture(u_image, v_texCoord + onePixel * vec2(-1,  0)) * u_kernel_2[3] +
        texture(u_image, v_texCoord + onePixel * vec2( 0,  0)) * u_kernel_2[4] +
        texture(u_image, v_texCoord + onePixel * vec2( 1,  0)) * u_kernel_2[5] +
        texture(u_image, v_texCoord + onePixel * vec2(-1,  1)) * u_kernel_2[6] +
        texture(u_image, v_texCoord + onePixel * vec2( 0,  1)) * u_kernel_2[7] +
        texture(u_image, v_texCoord + onePixel * vec2( 1,  1)) * u_kernel_2[8] ;

    // Divide the sum by the weight but just use rgb
    // we'll set alpha to 1.0
    color = colorSum;
}