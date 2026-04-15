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
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
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
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.bmc.app.data.Database
import com.bmc.app.models.ContentItem
import com.bmc.app.models.PathStep
import com.bmc.app.util.DeepLink
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PathDetailScreen(
    pathId: Int,
    pathTitle: String,
    onBack: () -> Unit = {}
) {
    val context = LocalContext.current
    var steps by remember { mutableStateOf<List<PathStep>>(emptyList()) }
    var stepContents by remember { mutableStateOf<Map<Int, List<ContentItem>>>(emptyMap()) }

    LaunchedEffect(pathId) {
        withContext(Dispatchers.IO) {
            val db = Database.getInstance(context)
            val loadedSteps = db.pathSteps(pathId)
            steps = loadedSteps

            val contentsMap = mutableMapOf<Int, List<ContentItem>>()
            for (step in loadedSteps) {
                contentsMap[step.id] = db.pathStepContents(step.id)
            }
            stepContents = contentsMap
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(pathTitle) },
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
        if (steps.isEmpty()) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(innerPadding),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = "暂无步骤",
                    style = MaterialTheme.typography.bodyLarge,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        } else {
            LazyColumn(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(innerPadding)
            ) {
                items(steps) { step ->
                    StepCard(
                        step = step,
                        contents = stepContents[step.id] ?: emptyList(),
                        onContentClick = { item ->
                            DeepLink.open(context, item.sourceUrl, item.sourcePlatform)
                        }
                    )
                }
            }
        }
    }
}

@Composable
private fun StepCard(
    step: PathStep,
    contents: List<ContentItem>,
    onContentClick: (ContentItem) -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 12.dp)
    ) {
        // Step header with day number
        Row(
            horizontalArrangement = Arrangement.spacedBy(12.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            // Day badge
            Surface(
                shape = CircleShape,
                color = MaterialTheme.colorScheme.primary
            ) {
                Text(
                    text = step.day,
                    style = MaterialTheme.typography.labelMedium,
                    fontWeight = FontWeight.Bold,
                    color = MaterialTheme.colorScheme.onPrimary,
                    modifier = Modifier.padding(horizontal = 10.dp, vertical = 6.dp)
                )
            }

            Text(
                text = step.title,
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.SemiBold
            )
        }

        // Editorial note
        if (step.note.isNotEmpty()) {
            Text(
                text = step.note,
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.padding(start = 46.dp, top = 6.dp)
            )
        }

        // Linked content items
        if (contents.isNotEmpty()) {
            Column(
                modifier = Modifier.padding(start = 46.dp, top = 8.dp),
                verticalArrangement = Arrangement.spacedBy(0.dp)
            ) {
                contents.forEach { item ->
                    ContentRow(
                        item = item,
                        onClick = { onContentClick(item) }
                    )
                    if (item != contents.last()) {
                        HorizontalDivider(modifier = Modifier.padding(vertical = 2.dp))
                    }
                }
            }
        }
    }

    HorizontalDivider()
}
