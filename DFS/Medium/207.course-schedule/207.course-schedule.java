class Solution {
    public boolean canFinish(int numCourses, int[][] prerequisites) {
        List<List<Integer>>graph=new ArrayList<>();
        int []indegree=new int[numCourses];
        for(int i=0;i<numCourses;i++){
            graph.add(new ArrayList<>());
        }

        for(int []pre:prerequisites){
            int courses=pre[0], prereq=pre[1];
            graph.get(prereq).add(courses);
            indegree[courses]++;
        }

        Queue<Integer> q = new ArrayDeque<>();
        for(int i=0;i<numCourses;i++){
            if(indegree[i]==0) q.offer(i);
        }

        int count=0;
        while(!q.isEmpty()){
            int cur=q.poll();
            count++;
            for(int next:graph.get(cur)){
                indegree[next]--;
                if(indegree[next]==0) q.offer(next);
            }
        }

        return count==numCourses;
    }
}