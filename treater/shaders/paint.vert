#version 410 core

layout (location = 0) in vec3 position;
layout (location = 1) in vec3 color;
layout (location = 2) in vec2 texCoord;

uniform mat4 u_matrix;

out vec3 ourColor;
out vec2 v_texCoord;

void main()
{
    gl_Position =/* u_matrix * */vec4(position, 1.0);
    ourColor = color;
    v_texCoord = texCoord;    // pass the texture coords on to the fragment shader
}