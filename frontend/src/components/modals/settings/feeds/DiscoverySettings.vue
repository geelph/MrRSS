<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhPlay, PhBinoculars, PhInfo } from '@phosphor-icons/vue';
import { ButtonControl, InfoBox, SettingGroup, SettingItem } from '@/components/settings';

const { t } = useI18n();

const emit = defineEmits<{
  'discover-all': [];
}>();

function handleDiscoverAll() {
  emit('discover-all');
}
</script>

<template>
  <SettingGroup :icon="PhBinoculars" :title="t('modal.feed.feedDiscovery')">
    <InfoBox :icon="PhInfo" :content="t('modal.discovery.discoveryLongRunningWarning')" />

    <!-- Discover All Feeds -->
    <SettingItem
      :icon="PhBinoculars"
      :title="t('modal.discovery.discoverAllFeeds')"
      :description="t('modal.discovery.discoverAllFeedsDesc')"
    >
      <ButtonControl
        :label="t('modal.discovery.startDiscovery')"
        :icon="PhPlay"
        type="secondary"
        @click="handleDiscoverAll"
      />
    </SettingItem>
  </SettingGroup>
</template>

<style scoped>
@reference "../../../../style.css";
</style>
