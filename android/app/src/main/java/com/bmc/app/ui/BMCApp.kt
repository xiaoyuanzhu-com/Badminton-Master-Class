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
                    navController.navigate("category/${category.id}/${category.name}")
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
            val categoryName = backStackEntry.arguments?.getString("categoryName") ?: ""
            CategoryScreen(
                categoryId = categoryId,
                categoryName = categoryName,
                onSubcategoryTap = { sub ->
                    navController.navigate("category/${sub.id}/${sub.name}")
                },
                onBack = { navController.popBackStack() }
            )
        }
    }
}
