plugins {
    id("com.android.application")
    id("org.jetbrains.kotlin.android")
    id("org.jetbrains.kotlin.plugin.compose")
}

android {
  namespace = "com.wandero.app"
  compileSdk = 37

    defaultConfig {
        applicationId = "com.wandero.app"
        minSdk = 26
        targetSdk = 37
        versionCode = 1
        versionName = "1.0"
    }

    buildFeatures { compose = true }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
    kotlinOptions { jvmTarget = "17" }
}

dependencies {
    implementation("androidx.core:core-ktx:1.17.0")
    implementation("androidx.activity:activity-compose:1.12.1")
    implementation("androidx.compose.runtime:runtime:1.9.2")
    implementation("androidx.compose.ui:ui:1.9.2")
    implementation("androidx.compose.ui:ui-graphics:1.9.2")
    implementation("androidx.compose.ui:ui-tooling-preview:1.9.2")
    implementation("androidx.compose.foundation:foundation:1.9.2")
    debugImplementation("androidx.compose.ui:ui-tooling:1.9.2")
}
