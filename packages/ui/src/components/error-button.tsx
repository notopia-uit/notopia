'use client'
import { ArrowLeft } from "lucide-react";
import { Button } from "./shadcn/button";

export default function ErrorButton() {
  return (
    <div
      className="
          mt-8 flex flex-1 flex-col items-center justify-center text-center
          xl:items-start xl:text-start
        "
    >
      <div className="mb-3 flex items-center gap-3">
        <span className="text-sm font-semibold">404</span>
      </div>
      <h1 className="mb-2 text-4xl font-bold">Page Not Found</h1>
      <p>Oops! The page you're trying to access doesn't exist.</p>

      <Button className="mt-8 h-9 cursor-pointer px-4 py-2">
        <ArrowLeft />
        <span>Go Back Home</span>
      </Button>
    </div>
  )
}
