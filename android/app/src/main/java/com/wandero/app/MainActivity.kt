package com.wandero.app

import android.os.Bundle
import android.os.Handler
import android.os.Looper
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.RowScope
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicText
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

private val AppSurface = Color(0xFFEDE8DF)
private val Panel = Color(0xFFF9F7F2)
private val Card = Color.White
private val MutedSurface = Color(0xFFF3F7F1)
private val TextPrimary = Color(0xFF17201A)
private val TextSecondary = Color(0xFF404942)
private val TextMuted = Color(0xFF6E756F)
private val Brand = Color(0xFF2F6B54)
private val Historical = Color(0xFF7A3F20)
private val Border = Color(0xFFD8D0C4)

private enum class Screen { Explore, Place, Walk, Ready, Navigation, Routes, RouteDetail, Favorites }

private data class Place(
  val title: String,
  val meta: String,
  val duration: String,
)

private data class Route(
  val title: String,
  val meta: String,
  val tag: String,
  val stops: List<String>,
)

private fun pluralPlaces(count: Int): String = when {
  count % 10 == 1 && count % 100 != 11 -> "$count место"
  count % 10 in 2..4 && count % 100 !in 12..14 -> "$count места"
  else -> "$count мест"
}

private fun walkSummary(stops: List<Place>): String = when (stops.size) {
  0 -> "Добавьте хотя бы одно место"
  1 -> "1 место · маршрут будет построен после выбора ещё одной точки"
  else -> "${pluralPlaces(stops.size)} · порядок можно менять"
}

class MainActivity : ComponentActivity() {
  override fun onCreate(savedInstanceState: Bundle?) {
    super.onCreate(savedInstanceState)
    setContent { WanderoApp() }
  }
}

@Composable
private fun WanderoApp() {
  var screen by remember { mutableStateOf(Screen.Explore) }
  var walkCount by remember { mutableIntStateOf(0) }
  var places by remember { mutableStateOf<List<Place>>(emptyList()) }
  var routes by remember { mutableStateOf<List<Route>>(emptyList()) }
  var apiError by remember { mutableStateOf<String?>(null) }
  var selectedPlace by remember { mutableStateOf<Place?>(null) }
  var selectedRoute by remember { mutableStateOf<Route?>(null) }
  var walkStops by remember { mutableStateOf<List<Place>>(emptyList()) }
  var query by remember { mutableStateOf("") }
  var completedStops by remember { mutableIntStateOf(0) }
  val visiblePlaces = places.filter { place ->
    val matchesQuery = query.isBlank() || "${place.title} ${place.meta}".contains(query, ignoreCase = true)
    matchesQuery
  }

  LaunchedEffect(Unit) {
    Thread {
      val repository = ApiRepository()
      val placesResult = repository.fetchPlaces()
      val routesResult = repository.fetchRoutes()
      Handler(Looper.getMainLooper()).post {
        places = placesResult.getOrElse { emptyList() }.map { place ->
          Place(place.title, place.description.ifBlank { "Описание отсутствует" }, "Время посещения неизвестно")
        }
        routes = routesResult.getOrElse { emptyList() }.map { route ->
          val walk = route.walkingMinutes?.let { "$it мин пешком" } ?: "Время не указано"
          val distance = route.distanceMeters?.let { " · ${it / 1000.0} км" } ?: ""
          Route(route.title, "$walk$distance · ${pluralPlaces(route.stops.size)}", "${route.completionCount} прошли", route.stops)
        }
        apiError = listOfNotNull(placesResult.exceptionOrNull(), routesResult.exceptionOrNull())
          .firstOrNull()?.message
      }
    }.start()
  }

  Box(Modifier.fillMaxSize().background(AppSurface)) {
    when (screen) {
      Screen.Explore -> {
        ExploreScreen(
          places = visiblePlaces,
          apiError = apiError,
          walkCount = walkCount,
          query = query,
          onQueryChange = { query = it },
          onPlace = { place ->
            selectedPlace = place
            screen = Screen.Place
          },
          onAdd = { place ->
            if (place !in walkStops) {
              walkStops = walkStops + place
              walkCount = walkStops.size
            }
          },
          onWalk = { screen = Screen.Walk },
        )
      }

      Screen.Place -> {
        PlaceDetail(
          place = selectedPlace,
          onBack = { screen = Screen.Explore },
          onAdd = {
            selectedPlace?.let { place ->
              if (place !in walkStops) walkStops = walkStops + place
              walkCount = walkStops.size
            }
            screen = Screen.Walk
          },
        )
      }

      Screen.Walk -> {
        WalkScreen(
          stops = walkStops,
          onReady = { if (walkStops.size >= 2) screen = Screen.Ready },
          onExplore = { screen = Screen.Explore },
        )
      }

      Screen.Ready -> {
        ReadyScreen(
          stops = walkStops,
          onBack = { screen = Screen.Walk },
          onStart = { completedStops = 0; screen = Screen.Navigation },
        )
      }

      Screen.Navigation -> {
        NavigationScreen(
          stops = walkStops,
          completedStops = completedStops,
          onCompleteStop = {
            if (completedStops + 1 >= walkStops.size) screen = Screen.Walk else completedStops += 1
          },
          onClose = { screen = Screen.Ready },
        )
      }

      Screen.Routes -> {
        RoutesScreen(routes = routes, apiError = apiError, onRoute = { route ->
          selectedRoute = route
          screen = Screen.RouteDetail
        })
      }

      Screen.RouteDetail -> {
        RouteDetail(
          route = selectedRoute,
          onBack = { screen = Screen.Routes },
          onStart = {
            selectedRoute?.let { route ->
              walkStops = route.stops.map { Place(it, "Описание отсутствует", "Время посещения неизвестно") }
              walkCount = walkStops.size
              screen = Screen.Ready
            }
          },
        )
      }

      Screen.Favorites -> FavoritesScreen(onExplore = { screen = Screen.Explore })
    }

    if (screen in listOf(Screen.Explore, Screen.Walk, Screen.Routes, Screen.Favorites)) {
      BottomNav(
        selected = screen,
        onSelect = { screen = it },
        modifier = Modifier.align(Alignment.BottomCenter),
      )
    }
  }
}

@Composable
private fun ExploreScreen(
  places: List<Place>,
  apiError: String?,
  walkCount: Int,
  query: String,
  onQueryChange: (String) -> Unit,
  onPlace: (Place) -> Unit,
  onAdd: (Place) -> Unit,
  onWalk: () -> Unit,
) {
  Box(Modifier.fillMaxSize().padding(bottom = 72.dp)) {
    CityMap(Modifier.fillMaxSize(), showRoute = false)
    Column(
      Modifier.statusBarsPadding().padding(horizontal = 16.dp, vertical = 10.dp),
      verticalArrangement = Arrangement.spacedBy(10.dp),
    ) {
      BrandLine()
      SearchField(query = query, onQueryChange = onQueryChange)
    }

    Column(
      Modifier
        .align(Alignment.BottomCenter)
        .fillMaxWidth()
        .shadow(18.dp, RoundedCornerShape(topStart = 24.dp, topEnd = 24.dp))
        .clip(RoundedCornerShape(topStart = 24.dp, topEnd = 24.dp))
        .background(Panel)
        .padding(top = 10.dp, bottom = 12.dp),
    ) {
      Spacer(
        Modifier
          .align(Alignment.CenterHorizontally)
          .width(42.dp)
          .height(4.dp)
          .clip(CircleShape)
          .background(Border),
      )
      Row(
        Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 12.dp),
        verticalAlignment = Alignment.CenterVertically,
      ) {
        Column(Modifier.weight(1f)) {
          WText("Рядом с вами", 20, FontWeight.Bold)
          WText(apiError ?: if (places.isEmpty()) "Места не найдены" else "${places.size} интересных мест", 13, color = TextMuted)
        }
        SmallAction("Моя прогулка · $walkCount", onWalk)
      }
      LazyRow(contentPadding = PaddingValues(horizontal = 16.dp), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
        items(places) { place -> PlaceCard(place, { onPlace(place) }, { onAdd(place) }) }
      }
    }
  }
}

@Composable
private fun PlaceDetail(
  place: Place?,
  onBack: () -> Unit,
  onAdd: () -> Unit,
) {
  Column(Modifier.fillMaxSize().background(Panel).navigationBarsPadding()) {
    Box(Modifier.fillMaxWidth().height(264.dp)) {
      ArchitectureArt(Modifier.fillMaxSize(), Historical)
      BackButton(onBack, Modifier.statusBarsPadding().padding(16.dp))
      Tag("Историческое место", Modifier.align(Alignment.BottomStart).padding(16.dp))
    }
    LazyColumn(
      Modifier.weight(1f),
      contentPadding = PaddingValues(20.dp),
      verticalArrangement = Arrangement.spacedBy(18.dp),
    ) {
      item {
        WText(place?.title ?: "Место не выбрано", 28, FontWeight.Bold)
        Spacer(Modifier.height(8.dp))
        WText(place?.meta ?: "Историческая справка отсутствует", 14, FontWeight.SemiBold, Brand)
      }
      item {
        WText("Почему стоит зайти", 18, FontWeight.Bold)
        Spacer(Modifier.height(8.dp))
        WText(place?.meta ?: "Историческая справка для этого места пока не добавлена.", 15, color = TextSecondary, lineHeight = 22)
      }
      item {
        Column(
          Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(12.dp))
            .background(MutedSurface)
            .padding(16.dp),
        ) {
          WText("Исторический контекст", 13, FontWeight.Bold, Historical)
          Spacer(Modifier.height(6.dp))
          WText(place?.meta ?: "Нет проверенного исторического описания.", 14, color = TextSecondary, lineHeight = 20)
        }
      }
    }
    PrimaryButton("Добавить в прогулку", onAdd, Modifier.padding(16.dp).fillMaxWidth())
  }
}

@Composable
private fun WalkScreen(
  stops: List<Place>,
  onReady: () -> Unit,
  onExplore: () -> Unit,
) {
  Column(
    Modifier
      .fillMaxSize()
      .background(Panel)
      .statusBarsPadding()
      .padding(bottom = 72.dp),
  ) {
    Header("Моя прогулка", walkSummary(stops))
    Column(Modifier.padding(horizontal = 16.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
      SettingRow("Старт", "Моё местоположение")
      SettingRow("Финиш", "Вернуться к старту")
    }
    WText("Остановки", 18, FontWeight.Bold, modifier = Modifier.padding(16.dp))
    LazyColumn(
      Modifier.weight(1f),
      contentPadding = PaddingValues(horizontal = 16.dp),
      verticalArrangement = Arrangement.spacedBy(10.dp),
    ) {
      items(stops) { place -> StopRow(stops.indexOf(place) + 1, place) }
      item { SecondaryButton("+ Добавить место", onExplore, Modifier.fillMaxWidth()) }
    }
    Column(
      Modifier
        .fillMaxWidth()
        .background(Card)
        .border(1.dp, Border)
        .padding(16.dp),
    ) {
      WText(if (stops.size < 2) "Нужно минимум 2 места" else "${pluralPlaces(stops.size)} готовы к построению", 16, FontWeight.Bold)
      WText(if (stops.size < 2) "Добавьте ещё место, чтобы построить маршрут." else "Дистанция и время появятся после построения.", 13, color = TextMuted)
      Spacer(Modifier.height(12.dp))
      PrimaryButton("Построить прогулку", onReady, Modifier.fillMaxWidth(), enabled = stops.size >= 2)
    }
  }
}

@Composable
private fun ReadyScreen(
  stops: List<Place>,
  onBack: () -> Unit,
  onStart: () -> Unit,
) {
  Column(Modifier.fillMaxSize().background(Panel).navigationBarsPadding()) {
    Box(Modifier.fillMaxWidth().weight(1.15f)) {
      CityMap(Modifier.fillMaxSize(), showRoute = true)
      BackButton(onBack, Modifier.statusBarsPadding().padding(16.dp))
      Tag("Маршрут готов", Modifier.align(Alignment.TopEnd).statusBarsPadding().padding(16.dp))
    }
    Column(
      Modifier
        .fillMaxWidth()
        .weight(.85f)
        .clip(RoundedCornerShape(topStart = 24.dp, topEnd = 24.dp))
        .background(Panel)
        .padding(20.dp),
      verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
      WText("Моя прогулка", 24, FontWeight.Bold)
      WText("${pluralPlaces(stops.size)} · маршрут готов", 14, FontWeight.SemiBold, Brand)
      Column(
        Modifier
          .fillMaxWidth()
          .clip(RoundedCornerShape(12.dp))
          .background(MutedSurface)
          .padding(14.dp),
      ) {
        WText("Следующая остановка", 12, FontWeight.Bold, TextMuted)
        WText(stops.firstOrNull()?.title ?: "Остановка не выбрана", 17, FontWeight.Bold)
        WText("Откройте точку на карте, чтобы начать путь.", 13, color = TextSecondary)
      }
      Spacer(Modifier.weight(1f))
      PrimaryButton("Начать прогулку", onStart, Modifier.fillMaxWidth())
    }
  }
}

@Composable
private fun NavigationScreen(
  stops: List<Place>,
  completedStops: Int,
  onCompleteStop: () -> Unit,
  onClose: () -> Unit,
) {
  val current = stops.getOrNull(completedStops)
  Box(Modifier.fillMaxSize()) {
    CityMap(Modifier.fillMaxSize(), showRoute = true)
    Row(
      Modifier.fillMaxWidth().statusBarsPadding().padding(16.dp),
      horizontalArrangement = Arrangement.SpaceBetween,
    ) {
      BackButton(onClose)
      Tag("${completedStops + 1} из ${stops.size}")
    }
    Column(
      Modifier
        .align(Alignment.BottomCenter)
        .fillMaxWidth()
        .clip(RoundedCornerShape(topStart = 24.dp, topEnd = 24.dp))
        .background(Panel)
        .navigationBarsPadding()
        .padding(20.dp),
      verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
      WText("Следующая остановка", 14, FontWeight.Bold, Brand)
      Row(verticalAlignment = Alignment.CenterVertically) {
        Box(Modifier.size(52.dp).clip(CircleShape).background(Brand), contentAlignment = Alignment.Center) {
          WText("Далее", 12, FontWeight.Bold, Color.White)
        }
        Spacer(Modifier.width(14.dp))
        WText(current?.title ?: "Маршрут завершён", 22, FontWeight.Bold, modifier = Modifier.weight(1f), lineHeight = 27)
      }
      WText(current?.duration ?: "Все остановки пройдены", 14, color = TextMuted)
      PrimaryButton(if (completedStops + 1 == stops.size) "Завершить прогулку" else "Я на месте", onCompleteStop, Modifier.fillMaxWidth())
    }
  }
}

@Composable
private fun RoutesScreen(routes: List<Route>, apiError: String?, onRoute: (Route) -> Unit) {
  Column(
    Modifier
      .fillMaxSize()
      .background(AppSurface)
      .statusBarsPadding()
      .padding(bottom = 72.dp),
  ) {
    Header("Готовые прогулки", "Маршруты от редакции Wandero")
    Spacer(Modifier.height(12.dp))
    LazyColumn(
      Modifier.fillMaxSize(),
      contentPadding = PaddingValues(horizontal = 16.dp, vertical = 4.dp),
      verticalArrangement = Arrangement.spacedBy(14.dp),
    ) {
      if (apiError != null) item { WText("Не удалось загрузить маршруты: $apiError", 14, color = Historical) }
      if (routes.isEmpty() && apiError == null) item { WText("Готовых маршрутов пока нет.", 14, color = TextMuted) }
      items(routes) { route -> RouteCard(route) { onRoute(route) } }
    }
  }
}

@Composable
private fun RouteDetail(
  route: Route?,
  onBack: () -> Unit,
  onStart: () -> Unit,
) {
  Column(Modifier.fillMaxSize().background(Panel).navigationBarsPadding()) {
    Box(Modifier.fillMaxWidth().height(260.dp)) {
      ArchitectureArt(Modifier.fillMaxSize(), Brand)
      BackButton(onBack, Modifier.statusBarsPadding().padding(16.dp))
      Tag("Выбор Wandero", Modifier.align(Alignment.BottomStart).padding(16.dp))
    }
    LazyColumn(Modifier.weight(1f), contentPadding = PaddingValues(20.dp), verticalArrangement = Arrangement.spacedBy(16.dp)) {
      item {
        WText(route?.title ?: "Маршрут не выбран", 28, FontWeight.Bold)
        Spacer(Modifier.height(6.dp))
        WText(route?.meta ?: "Нет данных о маршруте", 14, FontWeight.SemiBold, Brand)
      }
      item { CityMap(Modifier.fillMaxWidth().height(170.dp).clip(RoundedCornerShape(14.dp)), showRoute = true) }
      item {
        WText(
          "Маршрут состоит из выбранных мест. Данные о длительности и расстоянии предоставляет сервис.",
          15,
          color = TextSecondary,
          lineHeight = 22,
        )
      }
      items(route?.stops.orEmpty()) { title ->
        Row(verticalAlignment = Alignment.CenterVertically) {
          NumberDot((route?.stops?.indexOf(title) ?: 0) + 1)
          Spacer(Modifier.width(12.dp))
          WText(title, 16, FontWeight.SemiBold)
        }
      }
    }
    PrimaryButton("Повторить маршрут", onStart, Modifier.padding(16.dp).fillMaxWidth())
  }
}

@Composable
private fun BrandLine() {
  Row(verticalAlignment = Alignment.CenterVertically) {
    Box(Modifier.size(32.dp).clip(RoundedCornerShape(9.dp)).background(Brand), contentAlignment = Alignment.Center) {
      WText("W", 17, FontWeight.Bold, Color.White)
    }
    Spacer(Modifier.width(9.dp))
    WText("Wandero", 20, FontWeight.Bold)
  }
}

@Composable
private fun SearchField(query: String, onQueryChange: (String) -> Unit) {
  Row(
    Modifier
      .fillMaxWidth()
      .shadow(8.dp, RoundedCornerShape(14.dp))
      .clip(RoundedCornerShape(14.dp))
      .background(Card)
      .padding(14.dp),
    verticalAlignment = Alignment.CenterVertically,
  ) {
    WText("Поиск", 13, FontWeight.SemiBold, color = Brand)
    Spacer(Modifier.width(10.dp))
    BasicTextField(
      value = query,
      onValueChange = onQueryChange,
      modifier = Modifier.weight(1f),
      textStyle = TextStyle(color = TextPrimary, fontSize = 15.sp),
      singleLine = true,
      decorationBox = { inner ->
        if (query.isEmpty()) WText("Куда пойдём?", 15, color = TextMuted)
        inner()
      },
    )
  }
}

@Composable
private fun CityMap(
  modifier: Modifier,
  showRoute: Boolean,
) {
  Canvas(modifier.background(AppSurface)) {
    drawRect(
      Color(0xFFDCE7DD),
      topLeft = Offset(size.width * .05f, size.height * .07f),
      size =
        Size(
          size.width * .22f,
          size.height * .23f,
        ),
    )
    drawOval(
      Color(0xFFD7E4EA),
      topLeft = Offset(size.width * .60f, -size.height * .1f),
      size =
        Size(
          size.width * .48f,
          size.height * 1.25f,
        ),
    )
    val street = Color(0xFFF9F7F2)
    for (i in 0..7) {
      val x = size.width * (i + 1) / 9f
      drawLine(street, Offset(x, 0f), Offset(x * .78f, size.height), strokeWidth = 8f)
    }
    for (i in 0..10) {
      val y = size.height * (i + 1) / 12f
      drawLine(street, Offset(0f, y), Offset(size.width, y * 1.04f), strokeWidth = 7f)
    }
    if (showRoute) {
      val p =
        Path().apply {
          moveTo(size.width * .18f, size.height * .72f)
          cubicTo(size.width * .28f, size.height * .58f, size.width * .35f, size.height * .60f, size.width * .46f, size.height * .46f)
          cubicTo(size.width * .56f, size.height * .32f, size.width * .68f, size.height * .44f, size.width * .78f, size.height * .22f)
        }
      drawPath(p, Color.White, style = Stroke(12f, cap = StrokeCap.Round))
      drawPath(p, Brand, style = Stroke(7f, cap = StrokeCap.Round))
    }
    val pins = listOf(Offset(.18f, .72f), Offset(.46f, .46f), Offset(.78f, .22f), Offset(.30f, .27f), Offset(.66f, .66f))
    pins.forEachIndexed { index, point ->
      drawCircle(if (showRoute && index < 3) Brand else Historical, 15f, Offset(size.width * point.x, size.height * point.y))
      drawCircle(Color.White, 6f, Offset(size.width * point.x, size.height * point.y))
    }
  }
}

@Composable
private fun ArchitectureArt(
  modifier: Modifier,
  accent: Color,
) {
  Canvas(modifier.background(Color(0xFFD8D0C4))) {
    drawRect(Color(0xFFBFCED2), size = Size(size.width, size.height * .55f))
    drawCircle(Color(0xFFF0D79A), size.minDimension * .08f, Offset(size.width * .82f, size.height * .18f))
    drawRect(accent, Offset(size.width * .12f, size.height * .45f), Size(size.width * .76f, size.height * .5f))
    drawRect(Color(0xFFE7D9BD), Offset(size.width * .18f, size.height * .38f), Size(size.width * .64f, size.height * .12f))
    repeat(5) { i ->
      val x = size.width * (.22f + i * .14f)
      drawRect(Color(0xFFF5EBD6), Offset(x, size.height * .52f), Size(size.width * .055f, size.height * .34f))
    }
    drawArc(
      color = Color(0xFFF5EBD6),
      startAngle = 180f,
      sweepAngle = 180f,
      useCenter = true,
      topLeft = Offset(size.width * .34f, size.height * .23f),
      size = Size(size.width * .32f, size.height * .30f),
    )
  }
}

@Composable
private fun PlaceCard(
  place: Place,
  onClick: () -> Unit,
  onAdd: () -> Unit,
) {
  Column(
    Modifier
      .width(270.dp)
      .clip(RoundedCornerShape(14.dp))
      .background(Card)
      .clickable(onClick = onClick)
      .padding(12.dp),
    verticalArrangement = Arrangement.spacedBy(8.dp),
  ) {
    Box(Modifier.fillMaxWidth().height(82.dp).clip(RoundedCornerShape(9.dp))) { ArchitectureArt(Modifier.fillMaxSize(), Historical) }
    WText(place.title, 16, FontWeight.Bold, maxLines = 1)
    WText(place.meta, 12, color = TextMuted, maxLines = 1)
    Row(verticalAlignment = Alignment.CenterVertically) {
      WText(place.duration, 12, FontWeight.SemiBold, Brand, modifier = Modifier.weight(1f))
      SmallAction("+ В прогулку", onAdd)
    }
  }
}

@Composable
private fun RouteCard(
  route: Route,
  onClick: () -> Unit,
) {
  Row(
    Modifier
      .fillMaxWidth()
      .height(142.dp)
      .clip(RoundedCornerShape(14.dp))
      .background(Card)
      .clickable(onClick = onClick)
      .padding(12.dp),
    horizontalArrangement = Arrangement.spacedBy(14.dp),
  ) {
    Box(Modifier.width(116.dp).fillMaxHeight().clip(RoundedCornerShape(10.dp))) { ArchitectureArt(Modifier.fillMaxSize(), Brand) }
    Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(7.dp)) {
      WText(route.tag.uppercase(), 11, FontWeight.Bold, Brand)
      WText(route.title, 18, FontWeight.Bold, maxLines = 2)
      Spacer(Modifier.weight(1f))
      WText(route.meta, 12, color = TextMuted)
    }
  }
}

@Composable
private fun StopRow(
  number: Int,
  place: Place,
) {
  Row(
    Modifier
      .fillMaxWidth()
      .clip(
        RoundedCornerShape(12.dp),
      ).background(Card)
      .border(1.dp, Border, RoundedCornerShape(12.dp))
      .padding(12.dp),
    verticalAlignment = Alignment.CenterVertically,
  ) {
    NumberDot(number)
    Spacer(Modifier.width(12.dp))
    Column(Modifier.weight(1f)) {
      WText(place.title, 15, FontWeight.Bold)
      WText(place.duration, 12, color = TextMuted)
    }
  }
}

@Composable
private fun SettingRow(
  label: String,
  value: String,
) {
  Row(
    Modifier
      .fillMaxWidth()
      .clip(
        RoundedCornerShape(10.dp),
      ).background(Card)
      .border(1.dp, Border, RoundedCornerShape(10.dp))
      .padding(14.dp),
    verticalAlignment = Alignment.CenterVertically,
  ) {
    WText(label, 12, FontWeight.Bold, TextMuted, modifier = Modifier.width(58.dp))
    WText(value, 14, FontWeight.SemiBold, modifier = Modifier.weight(1f))
  }
}

@Composable
private fun Header(
  title: String,
  subtitle: String,
) {
  Column(Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 14.dp)) {
    BrandLine()
    Spacer(Modifier.height(18.dp))
    WText(title, 28, FontWeight.Bold)
    WText(subtitle, 14, color = TextMuted)
  }
}

@Composable
@Composable
private fun BottomNav(
  selected: Screen,
  onSelect: (Screen) -> Unit,
  modifier: Modifier = Modifier,
) {
  Row(
    modifier
      .fillMaxWidth()
      .background(Panel)
      .navigationBarsPadding()
      .height(72.dp)
      .padding(horizontal = 8.dp),
    verticalAlignment = Alignment.CenterVertically,
  ) {
    NavItem("Карта", selected == Screen.Explore) { onSelect(Screen.Explore) }
    NavItem("Маршруты", selected == Screen.Routes) { onSelect(Screen.Routes) }
    NavItem("Прогулка", selected == Screen.Walk) { onSelect(Screen.Walk) }
    NavItem("Избранное", selected == Screen.Favorites) { onSelect(Screen.Favorites) }
  }
}

@Composable
private fun RowScope.NavItem(
  label: String,
  active: Boolean,
  onClick: () -> Unit,
) {
  Column(
    Modifier
      .weight(1f)
      .clip(RoundedCornerShape(12.dp))
      .clickable(onClick = onClick)
      .padding(vertical = 8.dp),
    horizontalAlignment = Alignment.CenterHorizontally,
  ) {
    WText(label, 11, if (active) FontWeight.Bold else FontWeight.Normal, if (active) Brand else TextMuted)
  }
}

@Composable
private fun Chip(
  text: String,
  selected: Boolean = false,
  onClick: (() -> Unit)? = null,
) {
  Box(
    Modifier
      .clip(CircleShape)
      .background(if (selected) Brand else Card)
      .border(1.dp, if (selected) Brand else Border, CircleShape)
      .then(if (onClick != null) Modifier.clickable(onClick = onClick) else Modifier)
      .padding(horizontal = 14.dp, vertical = 9.dp),
  ) { WText(text, 13, FontWeight.SemiBold, if (selected) Color.White else TextSecondary) }
}

@Composable
private fun Tag(
  text: String,
  modifier: Modifier = Modifier,
) {
  Box(modifier.clip(CircleShape).background(Brand).padding(horizontal = 12.dp, vertical = 7.dp)) {
    WText(text, 12, FontWeight.Bold, Color.White)
  }
}

@Composable
private fun BackButton(
  onClick: () -> Unit,
  modifier: Modifier = Modifier,
) {
  Box(
    modifier
      .size(44.dp)
      .shadow(6.dp, CircleShape)
      .clip(CircleShape)
      .background(Card)
      .clickable(onClick = onClick),
    contentAlignment = Alignment.Center,
  ) { WText("Назад", 12, FontWeight.SemiBold, TextPrimary) }
}

@Composable
private fun NumberDot(number: Int) {
  Box(Modifier.size(30.dp).clip(CircleShape).background(Brand), contentAlignment = Alignment.Center) {
    WText(number.toString(), 13, FontWeight.Bold, Color.White)
  }
}

@Composable
private fun PrimaryButton(
  text: String,
  onClick: () -> Unit,
  modifier: Modifier = Modifier,
  enabled: Boolean = true,
) {
  Box(
    modifier
      .heightIn(
        min = 52.dp,
      ).clip(RoundedCornerShape(12.dp))
      .background(if (enabled) Brand else Border)
      .then(if (enabled) Modifier.clickable(onClick = onClick) else Modifier)
      .padding(horizontal = 18.dp),
    contentAlignment = Alignment.Center,
  ) { WText(text, 15, FontWeight.Bold, if (enabled) Color.White else TextMuted, align = TextAlign.Center) }
}

@Composable
private fun FavoritesScreen(onExplore: () -> Unit) {
  Column(
    Modifier.fillMaxSize().background(Panel).statusBarsPadding().padding(16.dp),
    verticalArrangement = Arrangement.spacedBy(14.dp),
  ) {
    Header("Избранное", "Сохраняйте места, чтобы вернуться к ним позже")
    WText("В избранном пока нет мест.", 15, color = TextMuted)
    SecondaryButton("Открыть карту", onExplore, Modifier.fillMaxWidth())
  }
}

@Composable
private fun SecondaryButton(
  text: String,
  onClick: () -> Unit,
  modifier: Modifier = Modifier,
) {
  Box(
    modifier
      .heightIn(
        min = 48.dp,
      ).clip(
        RoundedCornerShape(12.dp),
      ).border(1.dp, Brand, RoundedCornerShape(12.dp))
      .clickable(onClick = onClick)
      .padding(horizontal = 16.dp),
    contentAlignment = Alignment.Center,
  ) { WText(text, 14, FontWeight.Bold, Brand, align = TextAlign.Center) }
}

@Composable
private fun SmallAction(
  text: String,
  onClick: () -> Unit,
) {
  Box(
    Modifier
      .clip(CircleShape)
      .background(MutedSurface)
      .clickable(onClick = onClick)
      .padding(horizontal = 11.dp, vertical = 8.dp),
  ) {
    WText(text, 12, FontWeight.Bold, Brand)
  }
}

@Composable
private fun WText(
  text: String,
  size: Int,
  weight: FontWeight = FontWeight.Normal,
  color: Color = TextPrimary,
  modifier: Modifier = Modifier,
  maxLines: Int = Int.MAX_VALUE,
  align: TextAlign = TextAlign.Start,
  lineHeight: Int = size + 4,
) {
  BasicText(
    text = text,
    modifier = modifier,
    style = TextStyle(color = color, fontSize = size.sp, fontWeight = weight, textAlign = align, lineHeight = lineHeight.sp),
    maxLines = maxLines,
    overflow = TextOverflow.Ellipsis,
  )
}
