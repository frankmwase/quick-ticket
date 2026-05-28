import { useRef, useEffect } from 'react';
import { useFrame, useThree } from '@react-three/fiber';
import * as THREE from 'three';

/**
 * black Ops 1 menu style
 */
export function ParallaxCamera() {
  const { camera } = useThree();
  const mouse = useRef({ x: 0, y: 0 });
  const target = useRef({ x: 0, y: 0 });

  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      // Normalize to -1...1
      mouse.current.x = (e.clientX / window.innerWidth) * 2 - 1;
      mouse.current.y = (e.clientY / window.innerHeight) * 2 - 1;
    };
    window.addEventListener('mousemove', handleMouseMove);
    return () => window.removeEventListener('mousemove', handleMouseMove);
  }, []);

  useFrame(() => {
    // Smooth lerp towards mouse position
    target.current.x += (mouse.current.x - target.current.x) * 0.05;
    target.current.y += (mouse.current.y - target.current.y) * 0.05;

    // Apply parallax offset to camera position
    camera.position.x = target.current.x * 2.5;
    camera.position.y = 1.5 - target.current.y * 1.2;
    camera.position.z = 8;
    camera.lookAt(new THREE.Vector3(0, 0, -2));
  });

  return null;
}
