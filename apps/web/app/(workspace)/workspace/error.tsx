import { Button } from '@ui/components/shadcn/button';
import { ArrowLeft } from 'lucide-react';

const ErrorPage = () => {
  return (
    <div
      className="
      grid min-h-screen w-full
      xl:grid-cols-2
    "
    >
      <div className="flex flex-col p-16">
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
      </div>
      <div
        className="
        relative hidden
        xl:block
      "
      >
        <img
          src="https://ui.shadcn.com/placeholder.svg"
          alt="placeholder image"
          className="
            absolute inset-0 h-full object-cover
            dark:brightness-[0.95] dark:invert
          "
        />
      </div>
    </div>
  );
};

export default ErrorPage;
