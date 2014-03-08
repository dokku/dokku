'use strict';

module.exports = function (grunt) {

  require('load-grunt-tasks')(grunt);

  grunt.initConfig({

    compass: {
      options: {
        sassDir: 'static/styles',
        cssDir: 'static/.tmp/styles',
        generatedImagesDir: 'static/.tmp/images/generated',
        imagesDir: 'static/images',
        javascriptsDir: 'static/scripts',
        fontsDir: 'static/styles/fonts',
        importPath: 'static/components',
        httpImagesPath: '/images',
        httpGeneratedImagesPath: '/images/generated',
        httpFontsPath: '/styles/fonts',
        relativeAssets: false,
        assetCacheBuster: false,
        bundleExec: true,
        raw: 'Sass::Script::Number.precision = 10\n'
      },
      dist: {
        options: {
          // generatedImagesDir: '<%= yeoman.dist %>/static/images/generated'
        }
      },
      server: {
        options: {
          debugInfo: false
        }
      }
    }
  });

  grunt.registerTask('heroku', [
    'compass:server'
  ]);

};
