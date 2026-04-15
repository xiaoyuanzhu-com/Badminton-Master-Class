package com.bmc.app.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.ExpandLess
import androidx.compose.material.icons.filled.ExpandMore
import androidx.compose.material.icons.filled.Favorite
import androidx.compose.material.icons.filled.FavoriteBorder
import androidx.compose.material.icons.filled.Folder
import androidx.compose.material.icons.filled.Notes
import androidx.compose.material.icons.filled.OpenInNew
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.runtime.remember
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.unit.dp
import coil.compose.AsyncImage
import com.bmc.app.data.Database
import com.bmc.app.data.UserState
import com.bmc.app.models.Category
import com.bmc.app.models.ContentItem
import com.bmc.app.util.DeepLink
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CategoryScreen(
    categoryId: Int,
    categoryName: String,
    onSubcategoryTap: (Category) -> Unit,
    onBack: () -> Unit = {}
) {
    val context = LocalContext.current
    var subcategories by remember { mutableStateOf<List<Category>>(emptyList()) }
    var contents by remember { mutableStateOf<List<ContentItem>>(emptyList()) }

    LaunchedEffect(categoryId) {
        withContext(Dispatchers.IO) {
            val db = Database.getInstance(context)
            subcategories = db.categories(parentId = categoryId)
            contents = db.contents(categoryId = categoryId)
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(categoryName) },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(
                            imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                            contentDescription = "返回"
                        )
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.primaryContainer,
                    titleContentColor = MaterialTheme.colorScheme.onPrimaryContainer
                )
            )
        }
    ) { innerPadding ->
        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(innerPadding)
        ) {
            // Subcategories section
            if (subcategories.isNotEmpty()) {
                item {
                    SectionHeader("子分类")
                }
                items(subcategories) { sub ->
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .clickable { onSubcategoryTap(sub) }
                            .padding(horizontal = 16.dp, vertical = 12.dp),
                        horizontalArrangement = Arrangement.spacedBy(12.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text(
                            text = sub.icon,
                            style = MaterialTheme.typography.titleMedium
                        )
                        Text(
                            text = sub.name,
                            style = MaterialTheme.typography.bodyLarge,
                            modifier = Modifier.weight(1f)
                        )
                        if (sub.contentCount > 0) {
                            Text(
                                text = "${sub.contentCount} 个内容",
                                style = MaterialTheme.typography.labelSmall,
                                color = SecondaryText
                            )
                        }
                    }
                    HorizontalDivider()
                }
            }

            // Content section
            if (contents.isNotEmpty()) {
                item {
                    SectionHeader("内容")
                }
                items(contents) { item ->
                    ContentRow(
                        item = item,
                        onClick = {
                            DeepLink.open(context, item.sourceUrl, item.sourcePlatform)
                        }
                    )
                    HorizontalDivider()
                }
            }

            // Empty state
            if (subcategories.isEmpty() && contents.isEmpty()) {
                item {
                    Box(
                        modifier = Modifier
                            .fillParentMaxSize(),
                        contentAlignment = Alignment.Center
                    ) {
                        Column(horizontalAlignment = Alignment.CenterHorizontally) {
                            Icon(
                                imageVector = Icons.Default.Folder,
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
                                text = "该分类下还没有内容",
                                style = MaterialTheme.typography.bodyMedium,
                                color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.7f),
                                modifier = Modifier.padding(top = 4.dp)
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun SectionHeader(title: String) {
    Text(
        text = title,
        style = MaterialTheme.typography.titleSmall,
        color = MaterialTheme.colorScheme.onSurfaceVariant,
        modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp)
    )
}

@Composable
internal fun ContentRow(
    item: ContentItem,
    onClick: () -> Unit,
    showFavorite: Boolean = true
) {
    val context = LocalContext.current
    val userState = remember { UserState.getInstance(context) }
    val isFav = userState.isFavorite(item.id)
    var showEditorNotes by remember { mutableStateOf(false) }

    val platformActionText = when (item.sourcePlatform) {
        "bilibili" -> "在B站观看"
        "xiaohongshu" -> "在小红书查看"
        "douyin" -> "在抖音观看"
        "wechat" -> "在微信查看"
        "youtube" -> "在YouTube观看"
        else -> "打开链接"
    }

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .padding(horizontal = 16.dp, vertical = 12.dp)
    ) {
        Row(
            horizontalArrangement = Arrangement.spacedBy(12.dp),
            verticalAlignment = Alignment.Top
        ) {
            ContentThumbnail(thumbnailUrl = item.thumbnailUrl)

            Column(
                modifier = Modifier.weight(1f),
                verticalArrangement = Arrangement.spacedBy(6.dp)
            ) {
                Text(
                    text = item.title,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold
                )

                if (item.summary.isNotEmpty()) {
                    Text(
                        text = item.summary,
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                        maxLines = 2
                    )
                }

                // Metadata row: platform badge, category, author, difficulty, duration
                Row(
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    PlatformBadge(platform = item.sourcePlatform)

                    if (item.categoryName.isNotEmpty()) {
                        CategoryBadge(name = item.categoryName)
                    }

                    if (item.authorName.isNotEmpty()) {
                        Text(
                            text = item.authorName,
                            style = MaterialTheme.typography.labelSmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }

                    if (item.difficulty.isNotEmpty()) {
                        ContentDifficultyBadge(difficulty = item.difficulty)
                    }

                    if (item.duration.isNotEmpty()) {
                        Text(
                            text = item.duration,
                            style = MaterialTheme.typography.labelSmall,
                            color = SecondaryText
                        )
                    }
                }

                // External link indicator
                Row(
                    horizontalArrangement = Arrangement.spacedBy(4.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Icon(
                        imageVector = Icons.Default.OpenInNew,
                        contentDescription = null,
                        tint = SecondaryText,
                        modifier = Modifier.size(12.dp)
                    )
                    Text(
                        text = platformActionText,
                        style = MaterialTheme.typography.labelSmall,
                        color = SecondaryText
                    )
                }
            }

            if (showFavorite) {
                IconButton(
                    onClick = { userState.toggleFavorite(item.id) },
                    modifier = Modifier.size(36.dp)
                ) {
                    Icon(
                        imageVector = if (isFav) Icons.Default.Favorite else Icons.Default.FavoriteBorder,
                        contentDescription = if (isFav) "取消收藏" else "收藏",
                        tint = if (isFav) ErrorRed else MaterialTheme.colorScheme.onSurfaceVariant,
                        modifier = Modifier.size(20.dp)
                    )
                }
            }
        }

        // Editor's notes (expandable)
        if (item.editorNotes.isNotEmpty()) {
            Row(
                modifier = Modifier
                    .padding(start = 72.dp, top = 6.dp)
                    .clickable { showEditorNotes = !showEditorNotes },
                horizontalArrangement = Arrangement.spacedBy(4.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Icon(
                    imageVector = Icons.Default.Notes,
                    contentDescription = null,
                    tint = SecondaryText,
                    modifier = Modifier.size(14.dp)
                )
                Text(
                    text = "编辑笔记",
                    style = MaterialTheme.typography.labelSmall,
                    fontWeight = FontWeight.Medium,
                    color = SecondaryText
                )
                Icon(
                    imageVector = if (showEditorNotes) Icons.Default.ExpandLess else Icons.Default.ExpandMore,
                    contentDescription = null,
                    tint = SecondaryText,
                    modifier = Modifier.size(14.dp)
                )
            }

            androidx.compose.animation.AnimatedVisibility(visible = showEditorNotes) {
                Text(
                    text = item.editorNotes,
                    style = MaterialTheme.typography.labelSmall,
                    color = SecondaryText,
                    modifier = Modifier.padding(start = 72.dp, top = 4.dp, end = 16.dp)
                )
            }
        }
    }
}

@Composable
internal fun ContentDifficultyBadge(difficulty: String) {
    val (displayName, badgeColor) = when (difficulty) {
        "beginner" -> "入门" to SuccessGreen
        "intermediate" -> "进阶" to SecondaryText
        "advanced" -> "高级" to ErrorRed
        else -> difficulty to Color.Gray
    }

    Surface(
        shape = RoundedCornerShape(50),
        color = badgeColor.copy(alpha = 0.15f)
    ) {
        Text(
            text = displayName,
            style = MaterialTheme.typography.labelSmall,
            fontWeight = FontWeight.Medium,
            color = badgeColor,
            modifier = Modifier.padding(horizontal = 6.dp, vertical = 1.dp)
        )
    }
}

@Composable
internal fun ContentThumbnail(thumbnailUrl: String = "") {
    val shape = RoundedCornerShape(6.dp)
    if (thumbnailUrl.isNotEmpty()) {
        AsyncImage(
            model = thumbnailUrl,
            contentDescription = null,
            contentScale = ContentScale.Crop,
            modifier = Modifier
                .size(width = 60.dp, height = 45.dp)
                .clip(shape)
        )
    } else {
        Box(
            modifier = Modifier
                .size(width = 60.dp, height = 45.dp)
                .background(
                    color = MaterialTheme.colorScheme.surfaceVariant,
                    shape = shape
                ),
            contentAlignment = Alignment.Center
        ) {
            Icon(
                imageVector = Icons.Default.PlayArrow,
                contentDescription = null,
                tint = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.size(20.dp)
            )
        }
    }
}

@Composable
internal fun CategoryBadge(name: String) {
    Surface(
        shape = RoundedCornerShape(50),
        color = LightGray
    ) {
        Text(
            text = name,
            style = MaterialTheme.typography.labelSmall,
            fontWeight = FontWeight.Medium,
            color = SecondaryText,
            modifier = Modifier.padding(horizontal = 8.dp, vertical = 2.dp)
        )
    }
}

@Composable
internal fun PlatformBadge(platform: String) {
    val (displayName, badgeColor) = when (platform) {
        "bilibili" -> "B站" to Color(0xFFFB7299)
        "xiaohongshu" -> "小红书" to Color(0xFFFF2442)
        "douyin" -> "抖音" to Color(0xFF161823)
        "wechat" -> "微信" to Color(0xFF07C160)
        "youtube" -> "YouTube" to Color(0xFFFF0000)
        else -> "其他" to Color.Gray
    }

    Surface(
        shape = RoundedCornerShape(50),
        color = badgeColor.copy(alpha = 0.15f)
    ) {
        Text(
            text = displayName,
            style = MaterialTheme.typography.labelSmall,
            fontWeight = FontWeight.Medium,
            color = badgeColor,
            modifier = Modifier.padding(horizontal = 8.dp, vertical = 2.dp)
        )
    }
}
