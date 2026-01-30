class Solution {
    public int[] topKFrequent(int[] nums, int k) {
        Queue<int[]>pq=new PriorityQueue<>(Comparator.comparingInt(a -> a[1]));
        Map<Integer, Integer>freqMap=new HashMap<>();
        
        for(int num:nums) freqMap.put(num, freqMap.getOrDefault(num, 0)+1);

        freqMap.forEach((key, val)->{
            // System.out.println("key: "+key+", val: "+val);
            // pq.forEach(row->System.out.println("q key: "+row[0]+", ferq: "+row[1]));
            if(pq.size()<k){
                pq.offer(new int[]{key, val});
            }
            else if((val > freqMap.get(pq.peek()[0]))
            ){
                pq.poll();
                pq.offer(new int[]{key, val});
            }            
        });
        // pq.forEach(row->System.out.println("q key: "+row[0]+", ferq: "+row[1]));
        int []ans=new int[pq.size()];
        int index=0;
        for(int []row:pq) ans[index++]=row[0];
        return ans;
    }
}