uniform float uTime;
varying vec2 vUv;

void main() {
  // Vertical gradient fog with animated noise
  float fog = smoothstep(0.0, 1.0, vUv.y);
  float noise = sin(vUv.x * 40.0 + uTime) * sin(vUv.y * 30.0 - uTime * 0.5) * 0.02;
  float alpha = fog * 0.15 + noise;

  vec3 color = mix(vec3(0.0, 1.0, 0.25), vec3(0.0, 0.4, 0.1), vUv.y);
  gl_FragColor = vec4(color, alpha);
}
