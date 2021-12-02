const esbuild = require('esbuild');

(async () => {
  await esbuild.build({
    entryPoints: {
      app: 'src/index.ts',
    },
    bundle: true,
    outdir: '.',
    metafile: true,
  });
})();
