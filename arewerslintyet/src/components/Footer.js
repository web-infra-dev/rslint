import { Bundler, getBundler } from '../utils/bundler';

function FooterLink({ href, children }) {
  return (
    <a
      target="_blank"
      rel="noopener noreferrer"
      className="ml-2 mr-2"
      href={href}
    >
      {children}
    </a>
  );
}

export default function Footer() {
  const bundler = getBundler();
  return (
    <div className="text-center text-secondary-foreground mb-4">
      {bundler === Bundler.Turbopack ? (
        <>
          <FooterLink href="https://turbo.build/pack">
            Turbopack Docs
          </FooterLink>
          &middot;
          <FooterLink href="https://nextjs.org/blog/next-15">
            Next.js 15
          </FooterLink>
        </>
      ) : (
        <>
          <FooterLink href="https://rspack.dev/">Rspack Docs</FooterLink>
          &middot;
          <FooterLink href="https://nextjs.org/docs">Next.js Docs</FooterLink>
          &middot;
          <FooterLink href="https://areweturboyet.com/">
            Are We Turbo Yet?
          </FooterLink>
        </>
      )}
    </div>
  );
}
