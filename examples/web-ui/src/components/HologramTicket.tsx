import { useRef } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';

export function HologramTicket() {
  const groupRef = useRef<THREE.Group>(null);

  useFrame(({ clock }) => {
    if (!groupRef.current) return;
    groupRef.current.rotation.y = Math.sin(clock.getElapsedTime() * 0.5) * 0.3;
    groupRef.current.position.y = Math.sin(clock.getElapsedTime() * 0.8) * 0.15;
  });

  return (
    <group ref={groupRef} position={[3.5, 0.5, -2]}>
      {/* Ticket body */}
      <mesh>
        <boxGeometry args={[1.6, 1.0, 0.02]} />
        <meshStandardMaterial
          color="#00ff41"
          emissive="#00ff41"
          emissiveIntensity={0.3}
          transparent
          opacity={0.15}
          wireframe
        />
      </mesh>
      {/* Glow ring */}
      <mesh rotation={[Math.PI / 2, 0, 0]}>
        <torusGeometry args={[1.2, 0.01, 16, 64]} />
        <meshStandardMaterial
          color="#00ff41"
          emissive="#00ff41"
          emissiveIntensity={0.6}
          transparent
          opacity={0.3}
        />
      </mesh>
      {/* Inner wireframe cube */}
      <mesh>
        <boxGeometry args={[0.4, 0.4, 0.4]} />
        <meshStandardMaterial
          color="#ffb000"
          emissive="#ffb000"
          emissiveIntensity={0.5}
          wireframe
          transparent
          opacity={0.4}
        />
      </mesh>
    </group>
  );
}
