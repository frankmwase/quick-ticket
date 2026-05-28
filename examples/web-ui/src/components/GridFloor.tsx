import { useRef, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';

const gridVertexShader = `
varying vec2 vUv;
varying vec3 vPosition;
void main() {
  vUv = uv;
  vPosition = position;
  gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
}
`;

const gridFragmentShader = `
uniform float uTime;
uniform vec3 uColor;
varying vec2 vUv;
varying vec3 vPosition;
void main() {
  float gridX = abs(fract(vPosition.x * 0.5 - 0.5) - 0.5) / fwidth(vPosition.x * 0.5);
  float gridZ = abs(fract(vPosition.z * 0.5 - 0.5) - 0.5) / fwidth(vPosition.z * 0.5);
  float line = min(gridX, gridZ);
  float grid = 1.0 - min(line, 1.0);
  float dist = length(vPosition.xz) * 0.05;
  float fade = max(0.0, 1.0 - dist);
  float pulse = sin(vPosition.z * 0.3 + uTime * 2.0) * 0.5 + 0.5;
  vec3 color = uColor * grid * fade * (0.4 + pulse * 0.3);
  float alpha = grid * fade * 0.6;
  gl_FragColor = vec4(color, alpha);
}
`;

export function GridFloor() {
  const meshRef = useRef<THREE.Mesh>(null);

  const uniforms = useMemo(
    () => ({
      uTime: { value: 0 },
      uColor: { value: new THREE.Color(0x00ff41) },
    }),
    []
  );

  useFrame(({ clock }) => {
    uniforms.uTime.value = clock.getElapsedTime();
  });

  return (
    <mesh ref={meshRef} rotation={[-Math.PI / 2, 0, 0]} position={[0, -2, 0]}>
      <planeGeometry args={[60, 60, 1, 1]} />
      <shaderMaterial
        vertexShader={gridVertexShader}
        fragmentShader={gridFragmentShader}
        uniforms={uniforms}
        transparent
        side={THREE.DoubleSide}
        extensions={{ derivatives: true } as any}
      />
    </mesh>
  );
}
