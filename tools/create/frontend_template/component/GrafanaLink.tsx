import { GrafanaIcon } from "graph-drawing/src/GrafanaIcon"
import { FiExternalLink } from "react-icons/fi";

export interface GrafanaLinkProps {
    href: string
    target?: string
    icon?: React.ReactNode
    children?: React.ReactNode
}

export function IconLink({ href, target, icon, children }: GrafanaLinkProps) {
    return <a href={href} target={target}>
        {icon == null ? <FiExternalLink /> : icon}
        {children}
    </a>
}

export function GrafanaLink({ href, target, children }: GrafanaLinkProps) {
    return <IconLink href={href} target={target} icon={<GrafanaIcon />} children={children} />
}
