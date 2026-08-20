import { useI18n, useNavigate } from '@rspress/core/runtime';
// import { Hero as BaseHero } from '@rstack-dev/doc-ui/hero';
import { useI18nUrl } from './utils';
import styles from './Hero.module.scss';

const GITHUB_URL = 'https://github.com/web-infra-dev/rslint';
const LOGO_URL = 'https://assets.rspack.rs/rslint/rslint-logo.svg';

/**
 * Custom hero. Replaces the shared `@rstack-dev/doc-ui` <Hero> (commented
 * below): that component exposes no props/CSS-vars for size, alignment or
 * height, and its class names are build-hashed (not an API), so fitting it
 * into the side-by-side layout would have meant fragile selector overrides.
 * This owns its markup + module styles, faithfully replicating doc-ui's
 * visuals (gradient title, button styles, GitHub link, theme colours) and
 * only adjusting size + alignment for the hero/demo row. See Hero.module.scss.
 */
export function Hero() {
  const navigate = useNavigate();
  const tUrl = useI18nUrl();
  const t = useI18n<typeof import('i18n')>();
  const onClickGetStarted = () => {
    navigate(tUrl('/guide'));
  };

  return (
    <div className={styles.hero}>
      <img className={styles.logo} src={LOGO_URL} alt="Rslint" />
      <h1 className={styles.title}>Rslint</h1>
      <p className={styles.subtitle}>{t('subtitle')}</p>
      <p className={styles.description}>{t('slogan')}</p>
      <div className={styles.buttons}>
        <button
          type="button"
          className={`${styles.button} ${styles.buttonPrimary}`}
          onClick={onClickGetStarted}
        >
          {t('quickStart')}
        </button>
        <a
          className={`${styles.button} ${styles.buttonSecondary}`}
          href={GITHUB_URL}
          target="_blank"
          rel="noreferrer"
        >
          <svg
            className={styles.githubIcon}
            xmlns="http://www.w3.org/2000/svg"
            width="100%"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              fill="currentColor"
              d="M12 .297c-6.63 0-12 5.373-12 12c0 5.303 3.438 9.8 8.205 11.385c.6.113.82-.258.82-.577c0-.285-.01-1.04-.015-2.04c-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729c1.205.084 1.838 1.236 1.838 1.236c1.07 1.835 2.809 1.305 3.495.998c.108-.776.417-1.305.76-1.605c-2.665-.3-5.466-1.332-5.466-5.93c0-1.31.465-2.38 1.235-3.22c-.135-.303-.54-1.523.105-3.176c0 0 1.005-.322 3.3 1.23c.96-.267 1.98-.399 3-.405c1.02.006 2.04.138 3 .405c2.28-1.552 3.285-1.23 3.285-1.23c.645 1.653.24 2.873.12 3.176c.765.84 1.23 1.91 1.23 3.22c0 4.61-2.805 5.625-5.475 5.92c.42.36.81 1.096.81 2.22c0 1.606-.015 2.896-.015 3.286c0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"
            />
          </svg>
          GitHub
        </a>
      </div>
    </div>
  );
}

/* Previous shared-component implementation (kept for reference):
 *
 * import { Hero as BaseHero } from '@rstack-dev/doc-ui/hero';
 *
 * <BaseHero
 *   showStars={false}
 *   showOvalBg={false}
 *   onClickGetStarted={onClickGetStarted}
 *   title="Rslint"
 *   subTitle={t('subtitle')}
 *   description={t('slogan')}
 *   logoUrl={LOGO_URL}
 *   getStartedButtonText={t('quickStart')}
 *   githubURL={GITHUB_URL}
 * />
 */
