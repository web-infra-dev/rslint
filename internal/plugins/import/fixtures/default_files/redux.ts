declare function connect<T>(component: T): T;

const App = function () {
  return 1;
};

export default connect(App);
