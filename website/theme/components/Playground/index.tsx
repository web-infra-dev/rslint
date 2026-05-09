import { useEffect, useState } from "react";


export default function Playground() { 
  const [comp, setComp] = useState<React.FC<{}> | null>(null)
  useEffect(() => { 
    import('./Playground').then(({ default: Playground }) => { 
      setComp(Playground);
    })
  }, []);
  return<>{comp}</>
}


