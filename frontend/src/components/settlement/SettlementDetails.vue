<template>
  <el-dialog
    :model-value="modelValue"
    title="Settlement Details"
    @update:model-value="$emit('update:modelValue', $event)"
    width="700px"
  >
    <div v-if="settlement">
      <!-- Settlement Header -->
      <div class="bg-red-50 border-l-4 border-red-500 p-4 mb-6">
        <div class="flex items-center">
          <div class="flex-shrink-0">
            <el-icon :size="32" class="text-red-600"><Warning /></el-icon>
          </div>
          <div class="ml-3 flex-1">
            <h3 class="text-lg font-semibold text-red-900">
              Debt Settlement - {{ settlement.debtor.name }}
            </h3>
            <p class="text-sm text-red-700 mt-1">
              {{ formatDateTime(settlement.settlement_date) }}
            </p>
          </div>
        </div>
      </div>

      <!-- Summary Cards -->
      <div class="grid grid-cols-2 gap-4 mb-6">
        <div class="bg-gray-50 rounded-lg p-4">
          <dt class="text-sm font-medium text-gray-500">Original Debt</dt>
          <dd class="mt-1 flex items-baseline">
            <span class="text-2xl font-bold text-red-600">
              {{ settlement.original_debt_points }}
            </span>
            <span class="ml-2 text-sm text-gray-500">points</span>
          </dd>
        </div>
        <div class="bg-gray-50 rounded-lg p-4">
          <dt class="text-sm font-medium text-gray-500">Total Payment</dt>
          <dd class="mt-1 text-2xl font-bold text-gray-900">
            {{ formatVND(settlement.money_amount) }}
          </dd>
        </div>
      </div>

      <!-- Distribution Breakdown -->
      <div class="mb-6">
        <h4 class="text-sm font-semibold text-gray-900 mb-3">Distribution Breakdown</h4>
        <div class="space-y-3">
          <div class="flex items-center justify-between p-3 bg-green-50 rounded-lg">
            <div class="flex items-center gap-2">
              <el-icon class="text-green-600"><Wallet /></el-icon>
              <span class="font-medium text-green-900">Fund Contribution</span>
            </div>
            <span class="text-lg font-bold text-green-600">
              {{ formatVND(settlement.fund_amount) }}
            </span>
          </div>
          <div class="flex items-center justify-between p-3 bg-blue-50 rounded-lg">
            <div class="flex items-center gap-2">
              <el-icon class="text-blue-600"><Trophy /></el-icon>
              <span class="font-medium text-blue-900">Winner Distribution</span>
            </div>
            <span class="text-lg font-bold text-blue-600">
              {{ formatVND(settlement.winner_distribution) }}
            </span>
          </div>
        </div>
      </div>

      <!-- Winners List -->
      <div v-if="settlement.winners && settlement.winners.length > 0" class="mb-6">
        <h4 class="text-sm font-semibold text-gray-900 mb-3">
          Winners ({{ settlement.winners.length }})
        </h4>
        <div class="border rounded-lg overflow-hidden">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Winner</th>
                <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Money Received</th>
                <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Points Deducted</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-for="winner in settlement.winners" :key="winner.id">
                <td class="px-4 py-3 text-sm font-medium text-gray-900">
                  {{ winner.winner.name }}
                </td>
                <td class="px-4 py-3 text-sm text-right text-green-600 font-semibold">
                  {{ formatVND(winner.money_amount) }}
                </td>
                <td class="px-4 py-3 text-sm text-right text-red-600 font-semibold">
                  -{{ winner.points_deducted }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Process Explanation -->
      <div class="bg-blue-50 rounded-lg p-4">
        <h4 class="text-sm font-semibold text-blue-900 mb-2">
          <el-icon class="mr-1"><InfoFilled /></el-icon>
          Settlement Process
        </h4>
        <ol class="text-sm text-blue-800 space-y-1 ml-5 list-decimal">
          <li>{{ settlement.debtor.name }} reached {{ settlement.original_debt_points }} points (debt threshold)</li>
          <li>Total debt amount: {{ formatVND(settlement.money_amount) }}</li>
          <li>{{ formatVND(settlement.fund_amount) }} deposited to fund ({{ fundSplitPercent }}%)</li>
          <li>{{ formatVND(settlement.winner_distribution) }} distributed to {{ settlement.winners.length }} winner(s)</li>
          <li>Winners' points reduced proportionally</li>
          <li>{{ settlement.debtor.name }}'s score reset to 0</li>
        </ol>
      </div>
    </div>

    <template #footer>
      <el-button @click="$emit('update:modelValue', false)">Close</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { Warning, Wallet, Trophy, InfoFilled } from '@element-plus/icons-vue'
import type { DebtSettlement } from '@/types/settlement'
import { formatVND } from '@/utils/formatters'
import { formatDateTime } from '@/utils/date'

interface Props {
  modelValue: boolean
  settlement: DebtSettlement | null
  fundSplitPercent?: number
}

const props = withDefaults(defineProps<Props>(), {
  fundSplitPercent: 50
})

defineEmits<{
  'update:modelValue': [value: boolean]
}>()
</script>
