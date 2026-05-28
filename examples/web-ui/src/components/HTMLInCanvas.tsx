import { useRef, useState, useEffect } from 'react';
import { useFrame, useThree } from '@react-three/fiber';
import * as THREE from 'three';
import { createPortal } from 'react-dom';

export function HTMLInCanvas({ children }: { children: React.ReactNode }) {
  const { gl, camera, size } = useThree();
  const [container] = useState(() => {
    const div = document.createElement('div');
    div.id = 'hic-ui-container';
    div.style.width = '100%';
    div.style.height = '100%';
    div.style.position = 'absolute';
    div.style.top = '0';
    div.style.left = '0';
    // We must ensure the element is interactive
    div.style.pointerEvents = 'auto';
    return div;
  });

  const textureRef = useRef<THREE.Texture | null>(null);

  // Apply layoutsubtree and append container
  useEffect(() => {
    gl.domElement.setAttribute('layoutsubtree', '');
    gl.domElement.appendChild(container);

    return () => {
      if (gl.domElement.contains(container)) {
        gl.domElement.removeChild(container);
      }
      gl.domElement.removeAttribute('layoutsubtree');
    };
  }, [gl, container]);

  // Create texture
  useEffect(() => {
    const texture = new THREE.Texture();
    texture.minFilter = THREE.LinearFilter;
    texture.magFilter = THREE.LinearFilter;
    texture.format = THREE.RGBAFormat;
    // Essential for UI clarity
    texture.generateMipmaps = false;
    texture.colorSpace = THREE.SRGBColorSpace;
    textureRef.current = texture;

    return () => {
      texture.dispose();
    };
  }, []);

  useFrame(() => {
    if (!textureRef.current) return;
    const texture = textureRef.current;
    const ctx = gl.getContext() as any;

    // Use experimental HIC API if available
    if (ctx.texElementImage2D) {
      const texProps = gl.properties.get(texture) as any;
      let webglTexture = texProps.__webglTexture;
      
      if (!webglTexture) {
        webglTexture = ctx.createTexture();
        texProps.__webglTexture = webglTexture;
        texture.version++; // trigger three.js update
      }

      ctx.bindTexture(ctx.TEXTURE_2D, webglTexture);
      ctx.texElementImage2D(ctx.TEXTURE_2D, 0, ctx.RGBA, ctx.RGBA, ctx.UNSIGNED_BYTE, container);
      
      // Setup params on first frame
      if (texture.version === 1) {
        ctx.texParameteri(ctx.TEXTURE_2D, ctx.TEXTURE_WRAP_S, ctx.CLAMP_TO_EDGE);
        ctx.texParameteri(ctx.TEXTURE_2D, ctx.TEXTURE_WRAP_T, ctx.CLAMP_TO_EDGE);
        ctx.texParameteri(ctx.TEXTURE_2D, ctx.TEXTURE_MIN_FILTER, ctx.LINEAR);
        ctx.texParameteri(ctx.TEXTURE_2D, ctx.TEXTURE_MAG_FILTER, ctx.LINEAR);
        texture.version++;
      }
    }
  });

  // Calculate plane size to fit the screen exactly
  // Distance from camera where we place the UI plane
  const distance = 5;
  const vFov = (camera as THREE.PerspectiveCamera).fov * Math.PI / 180;
  const height = 2 * Math.tan(vFov / 2) * distance;
  const width = height * (size.width / size.height);

  return (
    <>
      {createPortal(children, container)}
      <mesh position={[0, 0, (camera.position.z - distance)]} renderOrder={999}>
        <planeGeometry args={[width, height]} />
        <meshBasicMaterial 
          map={textureRef.current} 
          transparent 
          depthTest={false} 
          depthWrite={false}
          opacity={1}
        />
      </mesh>
    </>
  );
}
