uniform float uTime;
uniform vec3 uColor;
varying vec2 vUv;
varying vec3 vPosition;

void main() {
  // Grid lines
  float gridX = abs(fract(vPosition.x * 0.5 - 0.5) - 0.5) / fwidth(vPosition.x * 0.5);
  float gridZ = abs(fract(vPosition.z * 0.5 - 0.5) - 0.5) / fwidth(vPosition.z * 0.5);
  float line = min(gridX, gridZ);
  float grid = 1.0 - min(line, 1.0);

  // Distance fade
  float dist = length(vPosition.xz) * 0.05;
  float fade = max(0.0, 1.0 - dist);

  // Scanning pulse
  float pulse = sin(vPosition.z * 0.3 + uTime * 2.0) * 0.5 + 0.5;

  vec3 color = uColor * grid * fade * (0.4 + pulse * 0.3);
  float alpha = grid * fade * 0.6;

  gl_FragColor = vec4(color, alpha);
}
