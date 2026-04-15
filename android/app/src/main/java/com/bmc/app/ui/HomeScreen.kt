package com.bmc.app.ui

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Clear
import androidx.compose.material.icons.filled.Inbox
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextField
import androidx.compose.material3.TextFieldDefaults
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import com.bmc.app.data.DataSync
import com.bmc.app.data.Database
import com.bmc.app.data.SyncState
import com.bmc.app.data.UserState
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.Surface
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import com.bmc.app.models.Category
import com.bmc.app.models.ContentItem
import com.bmc.app.models.LearningPath
import com.bmc.app.util.DeepLink
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HomeScreen(
    syncState: SyncState = SyncState.Idle,
    onCategoryTap: (Category) -> Unit,
    onPathTap: (LearningPath) -> Unit = {}
) {
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    val userState = remember { UserState.getInstance(context) }
    var categories by remember { mutableStateOf<List<Category>>(emptyList()) }
    var learningPaths by remember { mutableStateOf<List<LearningPath>>(emptyList()) }
    var favoriteContents by remember { mutableStateOf<List<ContentItem>>(emptyList()) }
    var searchQuery by remember { mutableStateOf("") }
    var searchResults by remember { mutableStateOf<List<ContentItem>>(emptyList()) }
    var isRefreshing by remember { mutableStateOf(false) }
    val isSearching = searchQuery.isNotBlank()

    // Read the snapshot so we recompose when favorites change
    val favoriteIds = userState.favorites.toList()

    LaunchedEffect(Unit) {
        withContext(Dispatchers.IO) {
            val db = Database.getInstance(context)
            categories = db.categories(parentId = null)
            learningPaths = db.learningPaths()
        }
    }

    // Reload favorite contents whenever the favorites list changes
    LaunchedEffect(favoriteIds) {
        if (favoriteIds.isEmpty()) {
            favoriteContents = emptyList()
        } else {
            favoriteContents = withContext(Dispatchers.IO) {
                Database.getInstance(context).contentsByIds(favoriteIds)
            }
        }
    }

    LaunchedEffect(searchQuery) {
        if (searchQuery.isBlank()) {
            searchResults = emptyList()
        } else {
            delay(300) // 300ms debounce — coroutine auto-cancelled on new keystroke
            searchResults = withContext(Dispatchers.IO) {
                Database.getInstance(context).searchContents(searchQuery)
            }
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("羽球大师课") },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.primaryContainer,
                    titleContentColor = MaterialTheme.colorScheme.onPrimaryContainer
                )
            )
        },
        bottomBar = {
            SyncStatusBar(syncState)
        }
    ) { innerPadding ->
        PullToRefreshBox(
            isRefreshing = isRefreshing,
            onRefresh = {
                scope.launch {
                    isRefreshing = true
                    DataSync.syncIfNeeded(context)
                    withContext(Dispatchers.IO) {
                        val db = Database.getInstance(context)
                        categories = db.categories(parentId = null)
                        learningPaths = db.learningPaths()
                        if (favoriteIds.isNotEmpty()) {
                            favoriteContents = db.contentsByIds(favoriteIds)
                        }
                    }
                    isRefreshing = false
                }
            },
            modifier = Modifier
                .fillMaxSize()
                .padding(innerPadding)
        ) {
            Column(modifier = Modifier.fillMaxSize()) {
            // Search bar
            TextField(
                value = searchQuery,
                onValueChange = { searchQuery = it },
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp, vertical = 8.dp),
                placeholder = { Text("搜索教程") },
                leadingIcon = {
                    Icon(Icons.Default.Search, contentDescription = "搜索")
                },
                trailingIcon = {
                    if (searchQuery.isNotEmpty()) {
                        IconButton(onClick = { searchQuery = "" }) {
                            Icon(Icons.Default.Clear, contentDescription = "清除")
                        }
                    }
                },
                singleLine = true,
                shape = RoundedCornerShape(12.dp),
                colors = TextFieldDefaults.colors(
                    focusedContainerColor = LightGray,
                    unfocusedContainerColor = LightGray,
                    focusedIndicatorColor = Color.Transparent,
                    unfocusedIndicatorColor = Color.Transparent,
                )
            )

            if (isSearching) {
                // Search results
                if (searchResults.isEmpty()) {
                    Box(
                        modifier = Modifier.fillMaxSize(),
                        contentAlignment = Alignment.Center
                    ) {
                        Text(
                            text = "无搜索结果",
                            style = MaterialTheme.typography.bodyLarge,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                } else {
                    LazyColumn(modifier = Modifier.fillMaxSize()) {
                        items(searchResults) { item ->
                            ContentRow(
                                item = item,
                                onClick = {
                                    DeepLink.open(context, item.sourceUrl, item.sourcePlatform)
                                }
                            )
                            HorizontalDivider()
                        }
                    }
                }
            } else {
                if (categories.isEmpty() && learningPaths.isEmpty() && favoriteContents.isEmpty()) {
                    Box(
                        modifier = Modifier.fillMaxSize(),
                        contentAlignment = Alignment.Center
                    ) {
                        Column(horizontalAlignment = Alignment.CenterHorizontally) {
                            Icon(
                                imageVector = Icons.Default.Inbox,
                                contentDescription = null,
                                modifier = Modifier.size(48.dp),
                                tint = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                            Text(
                                text = "暂无内容",
                                style = MaterialTheme.typography.bodyLarge,
                                color = MaterialTheme.colorScheme.onSurfaceVariant,
                                modifier = Modifier.padding(top = 12.dp)
                            )
                            Text(
                                text = "下拉刷新获取最新数据",
                                style = MaterialTheme.typography.bodyMedium,
                                color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.7f),
                                modifier = Modifier.padding(top = 4.dp)
                            )
                        }
                    }
                } else {
                    LazyColumn(modifier = Modifier.fillMaxSize()) {
                        // Favorites section
                        if (favoriteContents.isNotEmpty()) {
                            item {
                                Text(
                                    text = "我的收藏",
                                    style = MaterialTheme.typography.titleSmall,
                                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp)
                                )
                            }
                            items(favoriteContents) { item ->
                                ContentRow(
                                    item = item,
                                    onClick = {
                                        DeepLink.open(context, item.sourceUrl, item.sourcePlatform)
                                    }
                                )
                                HorizontalDivider()
                            }
                        }

                        // Learning paths section
                        if (learningPaths.isNotEmpty()) {
                            item {
                                LearningPathsSection(
                                    paths = learningPaths,
                                    userState = userState,
                                    onPathTap = onPathTap
                                )
                            }
                        }

                        // Categories section
                        if (categories.isNotEmpty()) {
                            item {
                                Text(
                                    text = "分类",
                                    style = MaterialTheme.typography.titleSmall,
                                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp)
                                )
                            }
                            items(categories) { category ->
                                Row(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .clickable { onCategoryTap(category) }
                                        .padding(horizontal = 16.dp, vertical = 14.dp),
                                    horizontalArrangement = Arrangement.spacedBy(12.dp),
                                    verticalAlignment = Alignment.CenterVertically
                                ) {
                                    Text(
                                        text = category.icon,
                                        style = MaterialTheme.typography.titleLarge
                                    )
                                    Text(
                                        text = category.name,
                                        style = MaterialTheme.typography.bodyLarge
                                    )
                                }
                                HorizontalDivider()
                            }
                        }
                    }
                }
            }
            }
        }
    }
}

@Composable
private fun LearningPathsSection(
    paths: List<LearningPath>,
    userState: UserState,
    onPathTap: (LearningPath) -> Unit
) {
    Column(modifier = Modifier.fillMaxWidth()) {
        Text(
            text = "学习路径",
            style = MaterialTheme.typography.titleSmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp)
        )

        LazyRow(
            contentPadding = PaddingValues(horizontal = 16.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            items(paths) { path ->
                LearningPathCard(
                    path = path,
                    userState = userState,
                    onClick = { onPathTap(path) }
                )
            }
        }

        HorizontalDivider(modifier = Modifier.padding(top = 12.dp))
    }
}

@Composable
private fun LearningPathCard(
    path: LearningPath,
    userState: UserState,
    onClick: () -> Unit
) {
    val completedCount = userState.pathProgress[path.id]?.size ?: 0
    val progress = if (path.stepCount > 0) completedCount.toFloat() / path.stepCount else 0f

    Card(
        modifier = Modifier
            .width(200.dp)
            .clickable(onClick = onClick),
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant
        )
    ) {
        Column(
            modifier = Modifier.padding(14.dp),
            verticalArrangement = Arrangement.spacedBy(6.dp)
        ) {
            Text(
                text = path.title,
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.SemiBold,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis
            )

            Text(
                text = path.summary,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis
            )

            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                Surface(
                    shape = RoundedCornerShape(50),
                    color = MaterialTheme.colorScheme.primary.copy(alpha = 0.1f)
                ) {
                    Text(
                        text = path.difficulty,
                        style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.primary,
                        modifier = Modifier.padding(horizontal = 8.dp, vertical = 2.dp)
                    )
                }

                Text(
                    text = "$completedCount/${path.stepCount} 完成",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }

            LinearProgressIndicator(
                progress = { progress },
                modifier = Modifier
                    .fillMaxWidth()
                    .height(4.dp),
                color = Color(0xFF007D48),
                trackColor = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.12f),
            )
        }
    }
}

@Composable
private fun SyncStatusBar(state: SyncState) {
    AnimatedVisibility(
        visible = state !is SyncState.Idle,
        enter = fadeIn(),
        exit = fadeOut()
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(vertical = 6.dp),
            horizontalArrangement = Arrangement.Center,
            verticalAlignment = Alignment.CenterVertically
        ) {
            when (state) {
                is SyncState.Syncing -> {
                    CircularProgressIndicator(
                        modifier = Modifier.size(14.dp),
                        strokeWidth = 2.dp,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                    Text(
                        text = "正在同步...",
                        style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                        modifier = Modifier.padding(start = 6.dp)
                    )
                }
                is SyncState.Success -> {
                    Text(
                        text = "已同步",
                        style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
                is SyncState.Failed -> {
                    Text(
                        text = "同步失败",
                        style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.error
                    )
                }
                else -> {}
            }
        }
    }
}
