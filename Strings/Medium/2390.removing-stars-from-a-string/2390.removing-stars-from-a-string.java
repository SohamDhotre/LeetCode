class Solution {
    public String removeStars(String s) {
        Deque<Character> que=new ArrayDeque<>();
        StringBuilder sb=new StringBuilder();
        for(char ch : s.toCharArray())
        {
            if(ch=='*' && !que.isEmpty()){
                que.removeLast();
            }else 
                que.addLast(ch);
        }
        for(char ch :que)
        {
            sb.append(ch);
        }
        return sb.toString();

    }
}