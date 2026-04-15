package com.bmc.app.ui

import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.platform.LocalContext
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.bmc.app.data.DataSync
import com.bmc.app.data.SyncState
import kotlinx.coroutines.delay
import java.net.URLDecoder
import java.net.URLEncoder

@Composable
fun BMCApp() {
    val navController = rememberNavController()
    val context = LocalContext.current
    val syncState by DataSync.state.collectAsState()

    // Trigger data sync on launch
    LaunchedEffect(Unit) {
        DataSync.syncIfNeeded(context)
    }

    // Auto-dismiss success/failed after a delay
    LaunchedEffect(syncState) {
        when (syncState) {
            is SyncState.Success -> { delay(2_000); DataSync.resetState() }
            is SyncState.Failed  -> { delay(3_000); DataSync.resetState() }
            else -> {}
        }
    }

    NavHost(navController = navController, startDestination = "home") {
        composable("home") {
            HomeScreen(
                syncState = syncState,
                onCategoryTap = { category ->
                    val encodedName = URLEncoder.encode(category.name, "UTF-8")
                    navController.navigate("category/${category.id}/${encodedName}")
                },
                onPathTap = { path ->
                    val encodedTitle = URLEncoder.encode(path.title, "UTF-8")
                    navController.navigate("path/${path.id}/${encodedTitle}")
                }
            )
        }
        composable(
            route = "category/{categoryId}/{categoryName}",
            arguments = listOf(
                navArgument("categoryId") { type = NavType.IntType },
                navArgument("categoryName") { type = NavType.StringType }
            )
        ) { backStackEntry ->
            val categoryId = backStackEntry.arguments?.getInt("categoryId") ?: return@composable
            val categoryName = URLDecoder.decode(
                backStackEntry.arguments?.getString("categoryName") ?: "", "UTF-8"
            )
            CategoryScreen(
                categoryId = categoryId,
                categoryName = categoryName,
                onSubcategoryTap = { sub ->
                    val encodedName = URLEncoder.encode(sub.name, "UTF-8")
                    navController.navigate("category/${sub.id}/${encodedName}")
                },
                onBack = { navController.popBackStack() }
            )
        }
        composable(
            route = "path/{pathId}/{pathTitle}",
            arguments = listOf(
                navArgument("pathId") { type = NavType.IntType },
                navArgument("pathTitle") { type = NavType.StringType }
            )
        ) { backStackEntry ->
            val pathId = backStackEntry.arguments?.getInt("pathId") ?: return@composable
            val pathTitle = URLDecoder.decode(
                backStackEntry.arguments?.getString("pathTitle") ?: "", "UTF-8"
            )
            PathDetailScreen(
                pathId = pathId,
                pathTitle = pathTitle,
                onBack = { navController.popBackStack() }
            )
        }
    }
}
