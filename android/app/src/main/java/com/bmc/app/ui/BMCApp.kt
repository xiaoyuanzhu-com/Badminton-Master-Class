package com.bmc.app.ui

import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.platform.LocalContext
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.bmc.app.data.DataSync

@Composable
fun BMCApp() {
    val navController = rememberNavController()
    val context = LocalContext.current

    // Trigger data sync on launch
    LaunchedEffect(Unit) {
        DataSync.syncIfNeeded(context)
    }

    NavHost(navController = navController, startDestination = "home") {
        composable("home") {
            HomeScreen(
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
                }
            )
        }
    }
}
