package com.wandero.app

import org.json.JSONArray
import java.net.HttpURLConnection
import java.net.URL

/** Development API client. Production builds must supply an HTTPS reverse-proxy URL. */
internal class ApiRepository(private val baseUrl: String = "http://10.0.2.2:8080") {
  fun fetchPlaces(): Result<List<ApiPlace>> = get("/api/locations") { payload ->
    val values = JSONArray(payload)
    List(values.length()) { index ->
      val item = values.getJSONObject(index)
      ApiPlace(
        id = item.getLong("id"),
        title = item.getString("name"),
        description = item.optString("description"),
      )
    }
  }

  fun fetchRoutes(): Result<List<ApiRoute>> = get("/api/routes") { payload ->
    val values = JSONArray(payload)
    List(values.length()) { index -> parseRoute(values.getJSONObject(index)) }
  }

  private fun parseRoute(item: org.json.JSONObject): ApiRoute {
    val stops = item.getJSONArray("stops")
    return ApiRoute(
      id = item.getLong("id"),
      title = item.getString("title"),
      description = item.optString("description"),
      walkingMinutes = item.optIntOrNull("walking_minutes"),
      distanceMeters = item.optIntOrNull("distance_meters"),
      completionCount = item.optInt("completion_count"),
      stops = List(stops.length()) { position -> stops.getJSONObject(position).getJSONObject("location").getString("name") },
    )
  }

  private fun <T> get(path: String, decode: (String) -> T): Result<T> = runCatching {
    val connection = (URL(baseUrl + path).openConnection() as HttpURLConnection).apply {
      requestMethod = "GET"
      connectTimeout = 10_000
      readTimeout = 10_000
      setRequestProperty("Accept", "application/json")
    }
    try {
      if (connection.responseCode !in 200..299) error("API вернул ошибку ${connection.responseCode}")
      val payload = connection.inputStream.bufferedReader().use { reader -> reader.readText() }
      decode(payload)
    } finally {
      connection.disconnect()
    }
  }
}

internal data class ApiPlace(val id: Long, val title: String, val description: String)
internal data class ApiRoute(
  val id: Long,
  val title: String,
  val description: String,
  val walkingMinutes: Int?,
  val distanceMeters: Int?,
  val completionCount: Int,
  val stops: List<String>,
)

private fun org.json.JSONObject.optIntOrNull(name: String): Int? =
  if (isNull(name) || !has(name)) null else getInt(name)
