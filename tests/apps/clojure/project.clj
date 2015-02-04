(defproject clojure-sample "1.0.1"
  :description "Hello World Clojure Web App"
  :dependencies [[org.clojure/clojure "1.6.0"]
                 [compojure "1.3.1"]
                 [ring/ring-jetty-adapter "1.3.2"]]
  :main ^:skip-aot sample.app
  :min-lein-version "2.0.0")
