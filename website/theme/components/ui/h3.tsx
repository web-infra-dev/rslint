import { ReactNode } from 'react';

export default function H3(props: { children: ReactNode; className?: string }) {
  return (
    <h3
      className={`scroll-m-20 text-2xl font-semibold tracking-tight ${props.className}`}
    >
      {props.children}
    </h3>
  );
}
