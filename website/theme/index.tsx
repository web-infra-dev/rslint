import { Layout as BaseLayout } from '@rspress/core/theme-original';
import { NavIcon } from '@rstackjs/doc-ui/nav-icon';
import { HomeLayout } from './pages';
import '@rstackjs/doc-ui/theme.css';
import './index.scss';

const Layout = () => {
  return <BaseLayout beforeNavTitle={<NavIcon />} />;
};

export { Layout, HomeLayout };
export * from '@rspress/core/theme-original';
