// `export = namespace` form — esModuleInterop synthesizes the default.
namespace MyNS {
  export const a = 1;
  export const b = 2;
}

export = MyNS;
