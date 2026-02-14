import { Hero } from "@/components/home/Hero";
import { Features } from "@/components/home/Features";
import { Installation } from "@/components/home/Installation";
import { Commands } from "@/components/home/Commands";

export default function Home() {
  return (
    <div className="overflow-x-hidden">
      <Hero />
      <Features />
      <Installation />
      <Commands />
    </div>
  );
}
