import './globals.css';
import { Open_Sans as OpenSans } from 'next/font/google';
import localFont from 'next/font/local';
import { ThemeProvider } from '@/components/theme-provider';
import { cn } from '@/lib/utils';
import { Linter, getLinter } from './bundler';

const geistFont = localFont({
  src: [
    {
      path: './Geist-Regular.woff2',
      weight: '400',
      style: 'normal',
    },
    {
      path: './Geist-SemiBold.woff2',
      weight: '600',
      style: 'normal',
    },
  ],
  display: 'swap',
});

const openSansFont = OpenSans({
  subsets: ['latin'],
  display: 'swap',
});

export const metadata = {
  title: 'Are We RSLint Yet?',
  description:
    'Progress towards 100% of lint tests passing for RSLint - A fast JavaScript and TypeScript linter written in Rust',
  icons: [
    {
      url: "data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='.9em' font-size='90'>🦀</text></svg>",
      type: 'image/svg+xml',
    },
  ],
};

export default function RootLayout({ children }) {
  const linter = getLinter();
  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={cn(
          geistFont.className,
          'bg-background text-foreground',
          'rslint',
        )}
      >
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <main id="root">{children}</main>
        </ThemeProvider>
      </body>
    </html>
  );
}
