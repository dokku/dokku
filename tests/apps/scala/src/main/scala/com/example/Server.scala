package com.example

import com.twitter.finagle.{Http, Service}
import com.twitter.util.{Await, Future}
import com.twitter.finagle.http.Response
import java.net.InetSocketAddress
import org.jboss.netty.handler.codec.http._
import util.Properties
import java.net.URI
import java.sql.Connection
import java.sql.DriverManager

object Server {
  def main(args: Array[String]) {
    val port = Properties.envOrElse("PORT", "8080").toInt
    println("Starting on port: "+port)

    val server = Http.serve(":" + port, new Hello)
    Await.ready(server)
  }
}

class Hello extends Service[HttpRequest, HttpResponse] {
  def apply(request: HttpRequest): Future[HttpResponse] = {
    if (request.getUri.endsWith("/db")) {
      showDatabase(request);
    } else {
      showHome(request);
    }
  }

  def showHome(request: HttpRequest): Future[HttpResponse] = {
    val response = Response()
    response.setStatusCode(200)
    response.setContentString("Hello from Scala!")
    Future(response)
  }

  def showDatabase(request: HttpRequest): Future[HttpResponse] = {
    val connection = getConnection
    val stmt = connection.createStatement
    stmt.executeUpdate("CREATE TABLE IF NOT EXISTS ticks (tick timestamp)")
    stmt.executeUpdate("INSERT INTO ticks VALUES (now())")

    val rs = stmt.executeQuery("SELECT tick FROM ticks")

    var out = ""
    while (rs.next) {
      out += "Read from DB: " + rs.getTimestamp("tick") + "\n"
    }

    val response = Response()
    response.setStatusCode(200)
    response.setContentString(out)
    Future(response)
  }

  def getConnection(): Connection = {
    val dbUri = new URI(System.getenv("DATABASE_URL"))
    val username = dbUri.getUserInfo.split(":")(0)
    val password = dbUri.getUserInfo.split(":")(1)
    val dbUrl = "jdbc:postgresql://" + dbUri.getHost + dbUri.getPath
    DriverManager.getConnection(dbUrl, username, password)
  }
}
