import { useEffect, useState } from "react";

// sideEffects
export default function Playground() { 
  if (import.meta.env.SSG_MD) {
    return `
# Playground

This is a rslint playground page
`
  }
  const [comp, setComp] = useState<React.FC<{}> | null>(null)
  useEffect(() => { 
    import('./Playground').then(({ default: Playground }) => { 
      setComp(Playground);
    })
  }, []);
  return <>{comp}</>
}


