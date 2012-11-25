#version 120
// Input vertex data, different for all executions of this shader.
// attribute vec3 vertexPosition_modelspace;

void main(){
	gl_Position = ftransform();
	gl_FrontColor = gl_Color * vec4(vec3(1), 0);
}