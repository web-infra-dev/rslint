import { Bundler, getBundler } from './bundler';

interface Props {
  title: string;
  description: string;
  percent: number;
  decision?: 'yes' | 'no' | null;
}

const RSPACK_WARNING = (
  <p className="max-w-xl italic mt-8 mb-8">
    next-rspack is currently experimental. It's not an official Next.js plugin,
    and is supported by the Rspack team in partnership with Next.js. Help
    improve Next.js and Rspack{' '}
    <a
      className="text-primary-foreground"
      href="https://github.com/vercel/next.js/discussions/77800"
    >
      by providing feedback
    </a>
    .
  </p>
);

export default function IsItReady({
  title,
  description,
  percent,
  decision: forcedDecision,
}: Props) {
  const decision =
    forcedDecision === 'yes'
      ? true
      : forcedDecision === 'no'
        ? false
        : percent === 100;

  return (
    <section className="text-center mb-4 flex flex-col items-center">
      {decision ? (
        <h1 className="text-4xl my-6 lg:text-6xl">
          {title}: <span className="font-semibold">YES</span>
          <i aria-hidden className="not-italic ml-4">
            {'\ud83c\udf89'}
          </i>
        </h1>
      ) : (
        <>
          <h1 className="text-4xl my-6 lg:text-6xl text-is-it-ready-no">
            {title}: <span className="font-semibold">NO</span>
          </h1>
          <p className="mt-2 mb-2">
            <span className="font-semibold">{percent}%</span> of RSLint{' '}
            {description} are passing though
            <i aria-hidden className="not-italic ml-1">
              {'\u2705'}
            </i>
          </p>
        </>
      )}
      {getBundler() === Bundler.Rspack && RSPACK_WARNING}
    </section>
  );
}
