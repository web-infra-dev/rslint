import { BackgroundImage } from '@rstack-dev/doc-ui/background-image';
import { containerStyle } from '@rstack-dev/doc-ui/section-style';
import { CopyRight } from '../components/Copyright';
import { Hero } from '../components/Hero';
import { LintDemo } from '../components/LintDemo';
import { Starfield } from '../components/LintDemo/Starfield';
import { ToolStack } from '../components/ToolStack';
import styles from './index.module.scss';

export function HomeLayout() {
  return (
    <div style={{ position: 'relative' }}>
      <Starfield />
      <BackgroundImage />
      {/* Same doc-ui section wrapper ToolStack uses, so the hero's gutters +
          max-width track Rstack's at every window size and the two sections
          line up. Its vertical padding also restores the hero's height. */}
      <section className={containerStyle}>
        <div className={styles.heroRow}>
          <Hero />
          <LintDemo />
        </div>
      </section>
      <ToolStack />
      <CopyRight />
    </div>
  );
}
