class Solution {
    public int lastStoneWeight(int[] stones) {
        Queue<Integer>pq=new PriorityQueue<>(Collections.reverseOrder());
        for(int stone:stones) pq.offer(stone);
        while(pq.size()>1){
            int stone1=pq.peek()!=null?pq.poll():0;
            int stone2=pq.peek()!=null?pq.poll():0;
            if(stone1==stone2) continue;            
            pq.add(Math.abs(stone1-stone2));
        }
        return pq.peek()!=null?pq.poll():0;
    }
}