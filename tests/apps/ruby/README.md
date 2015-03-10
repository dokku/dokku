# ruby-sample

This is a barebones Ruby app using the [Sinatra](http://www.sinatrarb.com) framework.

## Running Locally

Asumming you have [Ruby](https://www.ruby-lang.org), [Bundler](http://bundler.io) and [Heroku Toolbelt](https://toolbelt.heroku.com) installed on your machine:

```sh
git clone git@github.com:heroku/ruby-sample.git # or clone your own fork
cd ruby-sample
bundle
foreman start
```

Your app should now be running on [localhost:5000](http://localhost:5000/).

## Deploying to Heroku

```
heroku create
git push heroku master
heroku open
```

## Documentation

For more information about using Ruby on Heroku, see these Dev Center articles:

- [Ruby on Heroku](https://devcenter.heroku.com/categories/ruby)
- [Getting Started with Ruby on Heroku](https://devcenter.heroku.com/articles/getting-started-with-ruby)
- [Getting Started with Rails 4.x on Heroku](https://devcenter.heroku.com/articles/getting-started-with-rails4)
- [Heroku Ruby Support](https://devcenter.heroku.com/articles/ruby-support)
