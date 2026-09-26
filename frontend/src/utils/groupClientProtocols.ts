import type { ProtocolID } from '@/types'
import { protocolCatalog } from '@/api/admin/protocolCapabilities'

// 分组共用入口目录，账号原生能力在选号时检查。
function orderedProtocols(protocols: Iterable<ProtocolID>): ProtocolID[] {
  const selected = new Set(protocols)
  return (protocolCatalog.value?.protocols ?? []).filter(protocol => !protocol.upstream_only && selected.has(protocol.id)).map(protocol => protocol.id)
}

export function supportedGroupClientProtocols(): ProtocolID[] {
  return orderedProtocols(protocolCatalog.value?.groups[0]?.protocols ?? [])
}

export function defaultGroupClientProtocols(): ProtocolID[] {
  return orderedProtocols(protocolCatalog.value?.groups[0]?.defaults ?? [])
}

export function effectiveGroupClientProtocols(protocols: readonly ProtocolID[] | null | undefined): ProtocolID[] {
  if (!protocolCatalog.value) return [...(protocols ?? [])]
  const supported = new Set(protocolCatalog.value.groups[0]?.protocols ?? [])
  return orderedProtocols((protocols ?? []).filter(protocol => supported.has(protocol)))
}

export function hasGroupClientProtocol(protocols: readonly ProtocolID[], protocol: ProtocolID): boolean {
  return protocols.includes(protocol)
}

export function setGroupClientProtocol(protocols: readonly ProtocolID[], protocol: ProtocolID, enabled: boolean): ProtocolID[] {
  const next = new Set(protocols)
  if (enabled && supportedGroupClientProtocols().includes(protocol)) next.add(protocol)
  else next.delete(protocol)
  return effectiveGroupClientProtocols([...next])
}
