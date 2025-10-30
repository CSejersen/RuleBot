import "@/app/globals.css"
import { AppSidebar } from "@/components/layout/app-sidebar"
import { PageBreadcrumb } from "@/components/layout/page-breadcrumb"
import { Separator } from "@/components/ui/separator"
import { ThemeProvider } from "@/components/layout/theme-provider"
import { ModeToggle } from "@/components/layout/mode-toggle"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { Metadata } from "next"
import type { ReactNode } from "react"

export const metadata: Metadata = {
  title: "Home Automation Console",
  description: "Manage rules, devices, and integration",
}

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body>
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <SidebarProvider
            style={
              {
                "--sidebar-width": "19rem",
              } as React.CSSProperties
            }
          >
            <AppSidebar />

            <SidebarInset>
              {/* header */}
              <header className="flex h-16 shrink-0 items-center gap-2 px-4">
                <SidebarTrigger className="-ml-1" />
                <Separator
                  orientation="vertical"
                  className="mr-2 data-[orientation=vertical]:h-4"
                />
                <PageBreadcrumb />

                <div className="ml-auto">
                  <ModeToggle />
                </div>

              </header>

              <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
                {children}
              </div>
            </SidebarInset>
          </SidebarProvider>
        </ThemeProvider>
      </body>
    </html>
  )
}
