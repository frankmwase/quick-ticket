import { useRef } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';

export function HologramTicket() {
  const groupRef = useRef<THREE.Group>(null);
  const ring1Ref = useRef<THREE.Mesh>(null);
  const ring2Ref = useRef<THREE.Mesh>(null);
  const coreRef = useRef<THREE.Mesh>(null);

  useFrame(({ clock }) => {
    if (!groupRef.current || !ring1Ref.current || !ring2Ref.current || !coreRef.current) return;
    const time = clock.getElapsedTime();
    
    // Float the whole group
    groupRef.current.position.y = Math.sin(time * 0.8) * 0.15 + 0.5;
    
    // Rotate rings in opposite directions
    ring1Ref.current.rotation.x = Math.PI / 2 + Math.sin(time * 0.3) * 0.2;
    ring1Ref.current.rotation.y = time * 0.5;
    
    ring2Ref.current.rotation.x = Math.PI / 2 + Math.cos(time * 0.4) * 0.2;
    ring2Ref.current.rotation.y = -time * 0.7;

    // Rotate and pulse the core
    coreRef.current.rotation.y = time * 0.2;
    coreRef.current.rotation.x = time * 0.1;
    const scale = 1 + Math.sin(time * 2) * 0.05;
    coreRef.current.scale.set(scale, scale, scale);
  });

  return (
    <group ref={groupRef} position={[3.5, 0.5, -2]}>
      {/* Outer Glow ring */}
      <mesh ref={ring1Ref}>
        <torusGeometry args={[1.4, 0.015, 16, 64]} />
        <meshStandardMaterial
          color="#00f0ff"
          emissive="#00f0ff"
          emissiveIntensity={0.8}
          transparent
          opacity={0.4}
        />
      </mesh>
      
      {/* Inner fast ring */}
      <mesh ref={ring2Ref}>
        <torusGeometry args={[1.0, 0.008, 16, 64]} />
        <meshStandardMaterial
          color="#ffb000"
          emissive="#ffb000"
          emissiveIntensity={1.2}
          transparent
          opacity={0.6}
        />
      </mesh>

      {/* Vertical Data Cylinder */}
      <mesh>
        <cylinderGeometry args={[0.8, 0.8, 2.5, 16, 1, true]} />
        <meshStandardMaterial
          color="#00f0ff"
          emissive="#00f0ff"
          emissiveIntensity={0.2}
          transparent
          opacity={0.05}
          wireframe
        />
      </mesh>

      {/* Core Octahedron */}
      <mesh ref={coreRef}>
        <octahedronGeometry args={[0.5, 0]} />
        <meshStandardMaterial
          color="#e5e5e5"
          emissive="#00f0ff"
          emissiveIntensity={0.6}
          wireframe
          transparent
          opacity={0.8}
        />
      </mesh>

      {/* Floating data bits inside */}
      <mesh position={[0, 0.6, 0]}>
        <boxGeometry args={[0.1, 0.1, 0.1]} />
        <meshBasicMaterial color="#ffb000" />
      </mesh>
      <mesh position={[0, -0.6, 0]}>
        <boxGeometry args={[0.1, 0.1, 0.1]} />
        <meshBasicMaterial color="#ffb000" />
      </mesh>
    </group>
  );
}
